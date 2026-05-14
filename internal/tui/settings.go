package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type settingsModel struct {
	keys      keyMap
	configPath textinput.Model
	portStart  textinput.Model
	focus      int
	width      int
	err        string
}

func newSettings(s store.Settings) settingsModel {
	cp := textinput.New()
	cp.Placeholder = store.DefaultSSHConfigPath()
	cp.CharLimit = 256
	cp.Width = 50
	cp.SetValue(s.SSHConfigPath)
	cp.Focus()

	pp := textinput.New()
	pp.Placeholder = "9000"
	pp.CharLimit = 5
	pp.Width = 6
	pp.SetValue(strconv.Itoa(s.DefaultLocalPortStart))

	return settingsModel{
		keys:       defaultKeys(),
		configPath: cp,
		portStart:  pp,
		focus:      0,
	}
}

func (s settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "down":
			s.focus = (s.focus + 1) % 2
			s.applyFocus()
			return s, nil
		case "shift+tab", "up":
			s.focus = (s.focus + 1) % 2
			s.applyFocus()
			return s, nil
		}
	}
	var cmd tea.Cmd
	if s.focus == 0 {
		s.configPath, cmd = s.configPath.Update(msg)
	} else {
		s.portStart, cmd = s.portStart.Update(msg)
	}
	return s, cmd
}

func (s *settingsModel) applyFocus() {
	if s.focus == 0 {
		s.configPath.Focus()
		s.portStart.Blur()
	} else {
		s.portStart.Focus()
		s.configPath.Blur()
	}
}

func (s settingsModel) Build() (store.Settings, error) {
	port, err := strconv.Atoi(strings.TrimSpace(s.portStart.Value()))
	if err != nil || port <= 0 || port > 65535 {
		return store.Settings{}, fmt.Errorf("default port must be 1-65535")
	}
	return store.Settings{
		SSHConfigPath:         strings.TrimSpace(s.configPath.Value()),
		DefaultLocalPortStart: port,
	}, nil
}

func (s settingsModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Settings"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("SSH config path"))
	b.WriteString("\n")
	b.WriteString("  " + s.configPath.View())
	b.WriteString("  " + s.configPathHint())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Default local port start"))
	b.WriteString("\n")
	b.WriteString("  " + s.portStart.View())
	b.WriteString("\n\n")

	if s.err != "" {
		b.WriteString(errorStyle.Render(s.err) + "\n\n")
	}
	b.WriteString(hintStyle.Render("tab next field · ctrl+s save · esc cancel"))
	return frameStyle.Width(s.width - 2).Render(b.String())
}

func (s settingsModel) configPathHint() string {
	p := strings.TrimSpace(s.configPath.Value())
	if p == "" {
		p = store.DefaultSSHConfigPath()
	}
	if _, err := os.Stat(p); err == nil {
		return statusConnectedStyle.Render("✓ exists")
	}
	return statusErrorStyle.Render("✗ not found")
}
