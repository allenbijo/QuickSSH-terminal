package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	sshpkg "github.com/allenbijo/QuickSSH-terminal/internal/ssh"
	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type Model struct {
	store   *store.Store
	manager *sshpkg.Manager
	keys    keyMap

	schema store.Schema
	hosts  []store.SSHHost
	width  int
	height int

	screen   screen
	list     listModel
	editor   editorModel
	settings settingsModel
	confirm  confirmModel
	hostPick listModel // reused list model for host picker; here we use a simple cursor list

	hostPickItems []string
	hostPickCur   int

	statusCh chan sshpkg.StatusEvent
	flash    string
	flashErr string
}

func NewModel(s *store.Store) *Model {
	statusCh := make(chan sshpkg.StatusEvent, 64)
	m := &Model{
		store:    s,
		keys:     defaultKeys(),
		statusCh: statusCh,
	}
	m.manager = sshpkg.NewManager(s, func(ev sshpkg.StatusEvent) {
		// non-blocking emit
		select {
		case statusCh <- ev:
		default:
		}
	})
	m.list = newList(m.manager)
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadSchemaCmd(),
		m.loadHostsCmd(),
		m.waitStatusCmd(),
	)
}

func (m *Model) loadSchemaCmd() tea.Cmd {
	return func() tea.Msg {
		sc, err := m.store.Load()
		if err != nil {
			return errorMsg{err: err}
		}
		// Reset persisted status — we don't own those processes.
		for i := range sc.Aliases {
			sc.Aliases[i].Status = store.StatusDisconnected
			sc.Aliases[i].Enabled = false
		}
		return schemaLoadedMsg{schema: sc}
	}
}

func (m *Model) loadHostsCmd() tea.Cmd {
	return func() tea.Msg {
		sc, err := m.store.Load()
		path := ""
		if err == nil {
			path = sc.Settings.SSHConfigPath
		}
		hosts, err := m.manager.LoadHosts(path)
		return hostsLoadedMsg{hosts: hosts, err: err}
	}
}

func (m *Model) waitStatusCmd() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.statusCh
		if !ok {
			return nil
		}
		return statusMsg(ev)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		m.editor, _ = m.editor.Update(msg)
		m.settings, _ = m.settings.Update(msg)
		return m, cmd

	case schemaLoadedMsg:
		m.schema = msg.schema
		m.list.SetAliases(m.schema.Aliases)
		return m, nil

	case hostsLoadedMsg:
		if msg.err != nil {
			m.flashErr = msg.err.Error()
		} else {
			m.hosts = msg.hosts
		}
		return m, nil

	case statusMsg:
		// re-arm and refresh aliases from store
		return m, tea.Batch(m.waitStatusCmd(), m.loadSchemaCmd())

	case connectResultMsg:
		if msg.err != nil {
			m.flashErr = msg.err.Error()
		} else {
			m.flash = "connected"
		}
		return m, tea.Batch(m.loadSchemaCmd(), clearFlashCmd())

	case disconnectDoneMsg:
		return m, m.loadSchemaCmd()

	case sessionEndedMsg:
		return m, nil

	case errorMsg:
		if msg.err != nil {
			m.flashErr = msg.err.Error()
		}
		return m, clearFlashCmd()

	case clearFlashMsg:
		m.flash = ""
		m.flashErr = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit only on the list screen (so esc closes sub-screens cleanly).
	switch m.screen {
	case screenList:
		return m.handleListKey(msg)
	case screenEditor:
		return m.handleEditorKey(msg)
	case screenSettings:
		return m.handleSettingsKey(msg)
	case screenConfirm:
		return m.handleConfirmKey(msg)
	case screenHostPicker:
		return m.handleHostPickerKey(msg)
	}
	return m, nil
}

func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, m.quitCmd()
	case "n":
		m.openEditor(store.Alias{ID: uuid.NewString()}, true)
		return m, nil
	case "e":
		if a, ok := m.list.Selected(); ok {
			m.openEditor(a, false)
		}
		return m, nil
	case "d":
		if a, ok := m.list.Selected(); ok {
			m.confirm = newConfirm(confirmDelete, "Delete alias?", a.Name)
			m.screen = screenConfirm
		}
		return m, nil
	case "s":
		m.settings = newSettings(m.schema.Settings)
		m.screen = screenSettings
		return m, nil
	case "r":
		if a, ok := m.list.Selected(); ok {
			return m, m.refreshAliasCmd(a.ID)
		}
		return m, nil
	case " ", "enter":
		if a, ok := m.list.Selected(); ok {
			status := m.manager.AliasStatus(a.ID)
			if status == store.StatusConnected || status == store.StatusConnecting {
				return m, m.disconnectAliasCmd(a.ID)
			}
			return m, m.connectAliasCmd(a.ID)
		}
		return m, nil
	case "h":
		if a, ok := m.list.Selected(); ok {
			return m, m.openTerminalCmd(a)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenList
		return m, nil
	case "ctrl+s":
		alias, err := m.editor.Build()
		if err != nil {
			m.editor.err = err.Error()
			return m, nil
		}
		m.editor.err = ""
		m.screen = screenList
		m.flash = "saved \"" + alias.Name + "\""
		m.flashErr = ""
		return m, tea.Batch(m.saveAliasCmd(alias), clearFlashCmd())
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m *Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenList
		return m, nil
	case "ctrl+s":
		s, err := m.settings.Build()
		if err != nil {
			m.settings.err = err.Error()
			return m, nil
		}
		m.settings.err = ""
		m.screen = screenList
		m.flash = "settings saved"
		m.flashErr = ""
		return m, tea.Batch(m.saveSettingsCmd(s), clearFlashCmd())
	}
	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenList
		return m, nil
	case "enter":
		if m.confirm.choice {
			switch m.confirm.action {
			case confirmDelete:
				if a, ok := m.list.Selected(); ok {
					m.screen = screenList
					return m, m.deleteAliasCmd(a.ID)
				}
			case confirmQuit:
				m.screen = screenList
				return m, m.quitNowCmd()
			}
		}
		m.screen = screenList
		return m, nil
	}
	var cmd tea.Cmd
	m.confirm, cmd = m.confirm.Update(msg)
	return m, cmd
}

func (m *Model) handleHostPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.screen = screenList
		return m, nil
	case "up", "k":
		if m.hostPickCur > 0 {
			m.hostPickCur--
		}
	case "down", "j":
		if m.hostPickCur < len(m.hostPickItems)-1 {
			m.hostPickCur++
		}
	case "enter":
		if len(m.hostPickItems) == 0 {
			m.screen = screenList
			return m, nil
		}
		host := m.hostPickItems[m.hostPickCur]
		m.screen = screenList
		return m, m.runSSHCmd(host)
	}
	return m, nil
}

func (m *Model) openEditor(a store.Alias, isNew bool) {
	m.editor = newEditor(a, isNew, m.hosts, m.schema.Settings.DefaultLocalPortStart)
	m.screen = screenEditor
}

// ---------- commands ----------

func (m *Model) saveAliasCmd(alias store.Alias) tea.Cmd {
	return func() tea.Msg {
		_, err := m.store.Update(func(s *store.Schema) {
			idx := -1
			for i := range s.Aliases {
				if s.Aliases[i].ID == alias.ID {
					idx = i
					break
				}
			}
			if idx >= 0 {
				s.Aliases[idx] = alias
			} else {
				s.Aliases = append(s.Aliases, alias)
			}
		})
		if err != nil {
			return errorMsg{err: err}
		}
		return schemaReloadCmd(m)()
	}
}

func (m *Model) deleteAliasCmd(id string) tea.Cmd {
	return func() tea.Msg {
		_ = m.manager.DisconnectAlias(id)
		_, err := m.store.Update(func(s *store.Schema) {
			out := s.Aliases[:0]
			for _, a := range s.Aliases {
				if a.ID != id {
					out = append(out, a)
				}
			}
			s.Aliases = out
		})
		if err != nil {
			return errorMsg{err: err}
		}
		return schemaReloadCmd(m)()
	}
}

func (m *Model) saveSettingsCmd(settings store.Settings) tea.Cmd {
	return func() tea.Msg {
		_, err := m.store.Update(func(s *store.Schema) {
			s.Settings = settings
		})
		if err != nil {
			return errorMsg{err: err}
		}
		m.screen = screenList
		return tea.Batch(m.loadSchemaCmd(), m.loadHostsCmd())()
	}
}

func (m *Model) connectAliasCmd(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.manager.ConnectAlias(id)
		return connectResultMsg{aliasID: id, err: err}
	}
}

func (m *Model) disconnectAliasCmd(id string) tea.Cmd {
	return func() tea.Msg {
		_ = m.manager.DisconnectAlias(id)
		return disconnectDoneMsg{aliasID: id}
	}
}

func (m *Model) refreshAliasCmd(id string) tea.Cmd {
	return func() tea.Msg {
		_ = m.manager.DisconnectAlias(id)
		err := m.manager.ConnectAlias(id)
		return connectResultMsg{aliasID: id, err: err}
	}
}

func (m *Model) openTerminalCmd(a store.Alias) tea.Cmd {
	hosts := uniqueHosts(a)
	if len(hosts) == 0 {
		return func() tea.Msg { return errorMsg{err: errors.New("alias has no forwards")} }
	}
	if len(hosts) == 1 {
		return m.runSSHCmd(hosts[0])
	}
	m.hostPickItems = hosts
	m.hostPickCur = 0
	m.screen = screenHostPicker
	return nil
}

func (m *Model) runSSHCmd(host string) tea.Cmd {
	cmd := exec.Command("ssh", "-t", "-t", host)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return sessionEndedMsg{err: err}
	})
}

func (m *Model) quitCmd() tea.Cmd {
	if !m.hasActiveTunnels() {
		return m.quitNowCmd()
	}
	m.confirm = newConfirm(confirmQuit, "Quit?", "This will disconnect all active tunnels.")
	m.screen = screenConfirm
	return nil
}

func (m *Model) quitNowCmd() tea.Cmd {
	return func() tea.Msg {
		m.manager.DisconnectAll()
		return tea.Quit()
	}
}

func (m *Model) hasActiveTunnels() bool {
	for _, a := range m.schema.Aliases {
		s := m.manager.AliasStatus(a.ID)
		if s == store.StatusConnected || s == store.StatusConnecting {
			return true
		}
	}
	return false
}

func schemaReloadCmd(m *Model) tea.Cmd {
	return m.loadSchemaCmd()
}

func clearFlashCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}

func uniqueHosts(a store.Alias) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range a.Forwards {
		if !seen[f.SSHHost] && f.SSHHost != "" {
			seen[f.SSHHost] = true
			out = append(out, f.SSHHost)
		}
	}
	return out
}

// ---------- view ----------

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}
	header := m.renderHeader()
	footer := m.renderFooter()
	body := ""
	switch m.screen {
	case screenList:
		body = m.list.View()
	case screenEditor:
		body = m.editor.View()
	case screenSettings:
		body = m.settings.View()
	case screenConfirm:
		body = lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center, m.confirm.View())
	case screenHostPicker:
		body = m.renderHostPicker()
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

const bannerArt = `  /$$$$$$            /$$           /$$        /$$$$$$   /$$$$$$  /$$   /$$
 /$$__  $$          |__/          | $$       /$$__  $$ /$$__  $$| $$  | $$
| $$  \ $$ /$$   /$$ /$$  /$$$$$$$| $$   /$$| $$  \__/| $$  \__/| $$  | $$
| $$  | $$| $$  | $$| $$ /$$_____/| $$  /$$/|  $$$$$$ |  $$$$$$ | $$$$$$$$
| $$  | $$| $$  | $$| $$| $$      | $$$$$$/  \____  $$ \____  $$| $$__  $$
| $$/$$ $$| $$  | $$| $$| $$      | $$_  $$  /$$  \ $$ /$$  \ $$| $$  | $$
|  $$$$$$/|  $$$$$$/| $$|  $$$$$$$| $$ \  $$|  $$$$$$/|  $$$$$$/| $$  | $$
 \____ $$$ \______/ |__/ \_______/|__/  \__/ \______/  \______/ |__/  |__/
      \__/`

func (m *Model) renderHeader() string {
	active := 0
	total := len(m.schema.Aliases)
	for _, a := range m.schema.Aliases {
		s := m.manager.AliasStatus(a.ID)
		if s == store.StatusConnected || s == store.StatusConnecting {
			active++
		}
	}

	banner := bannerStyle.Render(bannerArt)
	bannerW := lipgloss.Width(banner)
	count := subtitleStyle.Render(fmt.Sprintf("%d active / %d", active, total))
	gap := m.width - bannerW - lipgloss.Width(count) - 2
	if gap < 1 {
		gap = 1
	}
	top := lipgloss.JoinHorizontal(lipgloss.Bottom, banner, strings.Repeat(" ", gap), count)
	return headerBarStyle.Width(m.width).Render(top)
}

func (m *Model) renderFooter() string {
	var hints []string
	switch m.screen {
	case screenList:
		hints = []string{"↑/↓", "select", "·", "⏎/space", "toggle", "·", "h", "open ssh", "·", "e", "edit", "·", "n", "new", "·", "d", "delete", "·", "s", "settings", "·", "r", "refresh", "·", "q", "quit"}
	case screenEditor:
		hints = []string{"tab", "next field", "·", "ctrl+a", "add forward", "·", "ctrl+x", "remove forward", "·", "ctrl+s", "save", "·", "esc", "cancel"}
	case screenSettings:
		hints = []string{"tab", "next field", "·", "ctrl+s", "save", "·", "esc", "cancel"}
	case screenConfirm:
		hints = []string{"←/→", "choose", "·", "⏎", "confirm", "·", "esc", "cancel"}
	case screenHostPicker:
		hints = []string{"↑/↓", "select", "·", "⏎", "open", "·", "esc", "cancel"}
	}
	line := strings.Join(hints, " ")
	if m.flashErr != "" {
		line = errorStyle.Render("✗ "+m.flashErr) + "   " + hintStyle.Render(line)
	} else if m.flash != "" {
		line = statusConnectedStyle.Render("✓ "+m.flash) + "   " + hintStyle.Render(line)
	} else {
		line = hintStyle.Render(line)
	}
	return footerBarStyle.Width(m.width).Render(line)
}

func (m *Model) renderHostPicker() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Pick an SSH host"))
	b.WriteString("\n\n")
	for i, h := range m.hostPickItems {
		cursor := glyphCursorEmpty
		style := valueStyle
		if i == m.hostPickCur {
			cursor = itemSelectedAccent.Render(glyphCursor)
			style = itemSelectedAccent
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(h)))
	}
	return frameStyle.Width(m.width - 4).Render(b.String())
}
