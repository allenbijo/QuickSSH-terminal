package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type forwardForm struct {
	id         string
	kind       store.ForwardKind
	localPort  textinput.Model
	remoteHost textinput.Model
	remotePort textinput.Model
	hostIndex  int
}

type editorModel struct {
	keys     keyMap
	alias    store.Alias
	isNew    bool
	name     textinput.Model
	forwards []forwardForm
	hosts    []store.SSHHost
	focus    int // 0 = name; 1..N = forward fields (4 per forward)
	width    int
	height   int
	err      string
	defaultPortStart int
}

func newEditor(alias store.Alias, isNew bool, hosts []store.SSHHost, defaultPortStart int) editorModel {
	keys := defaultKeys()
	name := textinput.New()
	name.Placeholder = "name"
	name.CharLimit = 64
	name.Width = 32
	name.SetValue(alias.Name)
	name.Focus()

	e := editorModel{
		keys:             keys,
		alias:            alias,
		isNew:            isNew,
		name:             name,
		hosts:            hosts,
		focus:            0,
		defaultPortStart: defaultPortStart,
	}
	for _, f := range alias.Forwards {
		e.forwards = append(e.forwards, newForwardForm(f, hosts))
	}
	if len(e.forwards) == 0 {
		e.forwards = append(e.forwards, newForwardForm(e.blankForward(), hosts))
	}
	return e
}

func (e editorModel) blankForward() store.PortForward {
	port := e.defaultPortStart
	used := map[int]bool{}
	for _, f := range e.forwards {
		if v, err := strconv.Atoi(f.localPort.Value()); err == nil {
			used[v] = true
		}
	}
	for used[port] {
		port++
	}
	host := ""
	if len(e.hosts) > 0 {
		host = e.hosts[0].Name
	}
	return store.PortForward{
		ID:         uuid.NewString(),
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: 8000,
		SSHHost:    host,
	}
}

func newForwardForm(f store.PortForward, hosts []store.SSHHost) forwardForm {
	lp := textinput.New()
	lp.Placeholder = "local"
	lp.CharLimit = 5
	lp.Width = 6
	lp.SetValue(strconv.Itoa(f.LocalPort))

	rh := textinput.New()
	rh.Placeholder = "remote host"
	rh.CharLimit = 64
	rh.Width = 22
	rh.SetValue(f.RemoteHost)

	rp := textinput.New()
	rp.Placeholder = "remote port"
	rp.CharLimit = 5
	rp.Width = 6
	rp.SetValue(strconv.Itoa(f.RemotePort))

	idx := 0
	for i, h := range hosts {
		if h.Name == f.SSHHost {
			idx = i
			break
		}
	}

	kind := f.Kind
	if kind == "" {
		kind = store.ForwardLocal
	}

	return forwardForm{
		id:         f.ID,
		kind:       kind,
		localPort:  lp,
		remoteHost: rh,
		remotePort: rp,
		hostIndex:  idx,
	}
}

// fieldVisible reports whether the global focus index points at a field the
// user can currently interact with. Dynamic (SOCKS) forwards hide the remote
// host and remote port fields.
func (e editorModel) fieldVisible(idx int) bool {
	if idx == 0 {
		return true
	}
	fIdx := (idx - 1) / 4
	field := (idx - 1) % 4
	if fIdx >= len(e.forwards) {
		return false
	}
	if e.forwards[fIdx].kind == store.ForwardDynamic && (field == 1 || field == 2) {
		return false
	}
	return true
}

// focusMove steps focus by dir (+1 / -1), skipping over any hidden fields.
func (e *editorModel) focusMove(dir int) {
	total := e.totalFields()
	if total == 0 {
		return
	}
	idx := e.focus
	for i := 0; i < total; i++ {
		idx += dir
		if idx < 0 {
			idx = total - 1
		}
		if idx >= total {
			idx = 0
		}
		if e.fieldVisible(idx) {
			e.focusField(idx)
			return
		}
	}
}

func (e editorModel) totalFields() int {
	return 1 + len(e.forwards)*4
}

func (e *editorModel) blurAll() {
	e.name.Blur()
	for i := range e.forwards {
		e.forwards[i].localPort.Blur()
		e.forwards[i].remoteHost.Blur()
		e.forwards[i].remotePort.Blur()
	}
}

func (e *editorModel) focusField(idx int) {
	e.blurAll()
	total := e.totalFields()
	if total == 0 {
		return
	}
	if idx < 0 {
		idx = total - 1
	}
	if idx >= total {
		idx = 0
	}
	e.focus = idx
	if idx == 0 {
		e.name.Focus()
		return
	}
	fIdx := (idx - 1) / 4
	field := (idx - 1) % 4
	if fIdx >= len(e.forwards) {
		return
	}
	f := &e.forwards[fIdx]
	switch field {
	case 0:
		f.localPort.Focus()
	case 1:
		f.remoteHost.Focus()
	case 2:
		f.remotePort.Focus()
	case 3:
		// host select — no textinput focus, handled by left/right
	}
}

func (e editorModel) Update(msg tea.Msg) (editorModel, tea.Cmd) {
	var cmds []tea.Cmd
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = m.Width
		e.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "tab":
			e.focusMove(1)
			return e, nil
		case "shift+tab":
			e.focusMove(-1)
			return e, nil
		case "ctrl+t":
			// Toggle the focused forward between local (-L) and SOCKS (-D).
			if e.focus > 0 {
				fIdx := (e.focus - 1) / 4
				if fIdx < len(e.forwards) {
					if e.forwards[fIdx].kind == store.ForwardDynamic {
						e.forwards[fIdx].kind = store.ForwardLocal
					} else {
						e.forwards[fIdx].kind = store.ForwardDynamic
					}
					// If the current field just got hidden, move to local port.
					if !e.fieldVisible(e.focus) {
						e.focusField(1 + fIdx*4)
					}
				}
			}
			return e, nil
		case "ctrl+a":
			e.forwards = append(e.forwards, newForwardForm(e.blankForward(), e.hosts))
			e.focusField(1 + (len(e.forwards)-1)*4)
			return e, nil
		case "ctrl+x":
			if e.focus > 0 {
				idx := (e.focus - 1) / 4
				if idx < len(e.forwards) {
					e.forwards = append(e.forwards[:idx], e.forwards[idx+1:]...)
					if len(e.forwards) == 0 {
						e.forwards = append(e.forwards, newForwardForm(e.blankForward(), e.hosts))
					}
					e.focusField(0)
				}
			}
			return e, nil
		}

		// host cycle on the sshHost subfield
		if e.focus > 0 {
			fIdx := (e.focus - 1) / 4
			field := (e.focus - 1) % 4
			if field == 3 && fIdx < len(e.forwards) && len(e.hosts) > 0 {
				switch m.String() {
				case "left", "h":
					e.forwards[fIdx].hostIndex = (e.forwards[fIdx].hostIndex - 1 + len(e.hosts)) % len(e.hosts)
					return e, nil
				case "right", "l", " ", "enter":
					e.forwards[fIdx].hostIndex = (e.forwards[fIdx].hostIndex + 1) % len(e.hosts)
					return e, nil
				}
			}
		}
	}

	// dispatch into the focused textinput
	if e.focus == 0 {
		var cmd tea.Cmd
		e.name, cmd = e.name.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		fIdx := (e.focus - 1) / 4
		field := (e.focus - 1) % 4
		if fIdx < len(e.forwards) {
			f := &e.forwards[fIdx]
			var cmd tea.Cmd
			switch field {
			case 0:
				f.localPort, cmd = f.localPort.Update(msg)
			case 1:
				f.remoteHost, cmd = f.remoteHost.Update(msg)
			case 2:
				f.remotePort, cmd = f.remotePort.Update(msg)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return e, tea.Batch(cmds...)
}

func (e editorModel) Build() (store.Alias, error) {
	name := strings.TrimSpace(e.name.Value())
	if name == "" {
		return store.Alias{}, fmt.Errorf("name is required")
	}
	if len(e.forwards) == 0 {
		return store.Alias{}, fmt.Errorf("at least one forward is required")
	}
	out := store.Alias{
		ID:       e.alias.ID,
		Name:     name,
		Enabled:  e.alias.Enabled,
		Status:   e.alias.Status,
		Forwards: make([]store.PortForward, 0, len(e.forwards)),
	}
	if out.ID == "" {
		out.ID = uuid.NewString()
	}
	if out.Status == "" {
		out.Status = store.StatusDisconnected
	}
	for i, f := range e.forwards {
		lp, err := strconv.Atoi(strings.TrimSpace(f.localPort.Value()))
		if err != nil || lp <= 0 || lp > 65535 {
			return store.Alias{}, fmt.Errorf("forward #%d: invalid local port", i+1)
		}
		host := ""
		if len(e.hosts) > 0 && f.hostIndex < len(e.hosts) {
			host = e.hosts[f.hostIndex].Name
		}
		if host == "" {
			return store.Alias{}, fmt.Errorf("forward #%d: pick an ssh host", i+1)
		}
		if f.kind == store.ForwardDynamic {
			out.Forwards = append(out.Forwards, store.PortForward{
				ID:        f.id,
				Kind:      store.ForwardDynamic,
				LocalPort: lp,
				SSHHost:   host,
			})
			continue
		}
		rp, err := strconv.Atoi(strings.TrimSpace(f.remotePort.Value()))
		if err != nil || rp <= 0 || rp > 65535 {
			return store.Alias{}, fmt.Errorf("forward #%d: invalid remote port", i+1)
		}
		out.Forwards = append(out.Forwards, store.PortForward{
			ID:         f.id,
			Kind:       store.ForwardLocal,
			LocalPort:  lp,
			RemoteHost: strings.TrimSpace(f.remoteHost.Value()),
			RemotePort: rp,
			SSHHost:    host,
		})
	}
	return out, nil
}

func (e editorModel) View() string {
	title := "New alias"
	if !e.isNew {
		title = "Edit alias"
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Name") + "  " + e.name.View())
	b.WriteString("\n\n")

	if len(e.hosts) == 0 {
		b.WriteString(errorStyle.Render("No SSH hosts found in your ~/.ssh/config — set the path in Settings (s)."))
		b.WriteString("\n\n")
	}

	b.WriteString(labelStyle.Render("Forwards") + "  " + hintStyle.Render("(tab: next field · ctrl+a: add · ctrl+x: remove · ctrl+t: local/SOCKS · ←/→ on host: cycle)"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Local forwards map a local port to one remote address. SOCKS proxies (ctrl+t) route any traffic through the SSH host."))
	b.WriteString("\n\n")

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render("    "),
		labelStyle.Render(padRightLbl("LOCAL PORT", 8)),
		labelStyle.Render("    "),
		labelStyle.Render(padRightLbl("REMOTE HOST", 24)),
		labelStyle.Render(padRightLbl("PORT", 8)),
		labelStyle.Render("    "),
		labelStyle.Render("VIA SSH HOST"),
	)
	b.WriteString(header + "\n")

	for i, f := range e.forwards {
		var row string
		if f.kind == store.ForwardDynamic {
			row = lipgloss.JoinHorizontal(
				lipgloss.Top,
				labelStyle.Render(fmt.Sprintf("%d.  ", i+1)),
				f.localPort.View(),
				itemDimStyle.Render("  ⇄ "),
				valueStyle.Render(padRightLbl("SOCKS proxy", 34)),
				e.renderHostSelect(i),
			)
		} else {
			row = lipgloss.JoinHorizontal(
				lipgloss.Top,
				labelStyle.Render(fmt.Sprintf("%d.  ", i+1)),
				f.localPort.View(),
				itemDimStyle.Render("  → "),
				f.remoteHost.View(),
				itemDimStyle.Render(" : "),
				f.remotePort.View(),
				itemDimStyle.Render("    "),
				e.renderHostSelect(i),
			)
		}
		focusBar := "  "
		if e.focus > 0 && (e.focus-1)/4 == i {
			focusBar = itemSelectedAccent.Render("│ ")
		}
		b.WriteString(focusBar + row + "\n")
	}
	b.WriteString("\n")
	b.WriteString(hintStyle.Render(
		"  Local:  local 9000 → remote 127.0.0.1 : 22  via bastion\n" +
			"          (connect to localhost:9000 to reach port 22 on the remote network)\n" +
			"  SOCKS:  local 1080  ⇄  via bastion\n" +
			"          (point your browser/app at socks5://localhost:1080 to tunnel all traffic)"))

	if e.err != "" {
		b.WriteString("\n" + errorStyle.Render(e.err))
	}
	b.WriteString("\n" + hintStyle.Render("ctrl+s save · esc cancel"))
	return frameStyle.Width(e.width - 2).Render(b.String())
}

func padRightLbl(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func (e editorModel) renderHostSelect(i int) string {
	if len(e.hosts) == 0 {
		return errorStyle.Render("<no hosts>")
	}
	name := e.hosts[e.forwards[i].hostIndex].Name
	box := "[ " + name + " ▾]"
	if e.focus > 0 && (e.focus-1)/4 == i && (e.focus-1)%4 == 3 {
		return itemSelectedAccent.Render(box)
	}
	return valueStyle.Render(box)
}

