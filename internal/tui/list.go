package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allenbijo/QuickSSH-terminal/internal/ssh"
	"github.com/allenbijo/QuickSSH-terminal/internal/store"
)

type listModel struct {
	keys    keyMap
	aliases []store.Alias
	manager *ssh.Manager
	cursor  int
	offset  int
	width   int
	height  int
}

func newList(m *ssh.Manager) listModel {
	return listModel{
		keys:    defaultKeys(),
		manager: m,
	}
}

func (l *listModel) SetAliases(a []store.Alias) {
	l.aliases = a
	if l.cursor >= len(l.aliases) {
		l.cursor = len(l.aliases) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

func (l listModel) Selected() (store.Alias, bool) {
	if len(l.aliases) == 0 || l.cursor < 0 || l.cursor >= len(l.aliases) {
		return store.Alias{}, false
	}
	return l.aliases[l.cursor], true
}

func (l listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = m.Width
		l.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if l.cursor > 0 {
				l.cursor--
			}
		case "down", "j":
			if l.cursor < len(l.aliases)-1 {
				l.cursor++
			}
		case "home", "g":
			l.cursor = 0
		case "end", "G":
			l.cursor = len(l.aliases) - 1
			if l.cursor < 0 {
				l.cursor = 0
			}
		}
	}
	l.clampOffset()
	return l, nil
}

func (l *listModel) clampOffset() {
	visible := l.visibleRows()
	if visible <= 0 {
		return
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+visible {
		l.offset = l.cursor - visible + 1
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

func (l listModel) visibleRows() int {
	// reserve ~13 lines for ASCII banner header + footer + help
	v := l.height - 14
	if v < 4 {
		v = 4
	}
	return v
}

func (l listModel) View() string {
	if len(l.aliases) == 0 {
		return l.renderEmpty()
	}

	var rows []string
	visible := l.visibleRows()
	end := l.offset + visible
	if end > len(l.aliases) {
		end = len(l.aliases)
	}
	for i := l.offset; i < end; i++ {
		rows = append(rows, l.renderRow(i))
	}
	return strings.Join(rows, "\n")
}

func (l listModel) renderEmpty() string {
	msg := lipgloss.JoinVertical(
		lipgloss.Center,
		itemDimStyle.Render("No aliases yet."),
		"",
		valueStyle.Render("Press ")+itemSelectedAccent.Render("n")+valueStyle.Render(" to create one,"),
		valueStyle.Render("or ")+itemSelectedAccent.Render("s")+valueStyle.Render(" to set your SSH config path."),
	)
	return lipgloss.Place(l.width-4, l.visibleRows(), lipgloss.Center, lipgloss.Center, msg)
}

func (l listModel) renderRow(i int) string {
	a := l.aliases[i]
	status := l.aliasStatus(a)
	glyph, glyphStyle := statusGlyph(status)

	connStr := ""
	if status == store.StatusConnected || status == store.StatusConnecting || status == store.StatusError {
		c, total := l.manager.ConnectedCount(a.ID)
		if total > 0 {
			connStr = fmt.Sprintf("%d/%d", c, total)
		}
	}

	preview := ""
	if len(a.Forwards) > 0 {
		f := a.Forwards[0]
		if f.IsDynamic() {
			preview = fmt.Sprintf("%d ⇄ SOCKS", f.LocalPort)
		} else {
			preview = fmt.Sprintf("%d → %s:%d", f.LocalPort, f.RemoteHost, f.RemotePort)
		}
		if len(a.Forwards) > 1 {
			preview += itemDimStyle.Render(fmt.Sprintf(" +%d", len(a.Forwards)-1))
		}
	}

	cursor := glyphCursorEmpty
	nameStyle := itemNameStyle
	if i == l.cursor {
		cursor = itemSelectedAccent.Render(glyphCursor)
		nameStyle = itemSelectedAccent
	}

	name := nameStyle.Render(padRight(a.Name, 18))
	g := glyphStyle.Render(glyph)
	conn := itemDetailStyle.Render(padLeft(connStr, 4))
	preview = itemDetailStyle.Render(preview)

	row := fmt.Sprintf("%s %s  %s  %s   %s", cursor, name, g, conn, preview)

	if i == l.cursor {
		errs := l.manager.Errors(a.ID)
		if len(errs) > 0 {
			row += "\n  " + itemSelectedBg.Render(errorStyle.Render("⚠ "+truncate(errs[0], l.width-8)))
		}
	}
	return row
}

func (l listModel) aliasStatus(a store.Alias) store.AliasStatus {
	return l.manager.AliasStatus(a.ID)
}

func statusGlyph(s store.AliasStatus) (string, lipgloss.Style) {
	switch s {
	case store.StatusConnected:
		return glyphConnected, statusConnectedStyle
	case store.StatusConnecting:
		return glyphConnecting, statusConnectingStyle
	case store.StatusError:
		return glyphError, statusErrorStyle
	default:
		return glyphDisconnected, statusDisconnectedStyle
	}
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padLeft(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return strings.Repeat(" ", n-len(s)) + s
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return s[:n-1] + "…"
}
