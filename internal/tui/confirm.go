package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmAction int

const (
	confirmDelete confirmAction = iota
	confirmQuit
)

type confirmModel struct {
	action  confirmAction
	prompt  string
	subject string
	choice  bool // true = yes
}

func newConfirm(action confirmAction, prompt, subject string) confirmModel {
	return confirmModel{
		action:  action,
		prompt:  prompt,
		subject: subject,
		choice:  false,
	}
}

func (c confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "left", "h", "y", "Y":
			c.choice = true
		case "right", "l", "n", "N":
			c.choice = false
		}
	}
	return c, nil
}

func (c confirmModel) View() string {
	yes := "[ Yes ]"
	no := "[ No ]"
	if c.choice {
		yes = itemSelectedAccent.Render("[ Yes ]")
	} else {
		no = itemSelectedAccent.Render("[ No ]")
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(c.prompt))
	b.WriteString("\n\n")
	if c.subject != "" {
		b.WriteString(valueStyle.Render(c.subject))
		b.WriteString("\n\n")
	}
	b.WriteString("    " + yes + "   " + no)
	b.WriteString("\n\n" + hintStyle.Render("←/→ choose · ⏎ confirm · esc cancel"))
	return dialogStyle.Render(b.String())
}
