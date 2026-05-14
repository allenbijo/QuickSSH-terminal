package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Open     key.Binding
	Edit     key.Binding
	New      key.Binding
	Delete   key.Binding
	Settings key.Binding
	Refresh  key.Binding
	Save     key.Binding
	Cancel   key.Binding
	AddItem  key.Binding
	RemoveItem key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Quit     key.Binding
	Help     key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Toggle:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		Open:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "open ssh")),
		Edit:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		New:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Settings:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Save:       key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		AddItem:    key.NewBinding(key.WithKeys("+", "ctrl+n"), key.WithHelp("+", "add forward")),
		RemoveItem: key.NewBinding(key.WithKeys("-", "ctrl+x"), key.WithHelp("-", "remove forward")),
		Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		ShiftTab:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}
}

func (k keyMap) listShort() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Toggle, k.Edit, k.New, k.Delete, k.Settings, k.Refresh, k.Quit}
}

func (k keyMap) editorShort() []key.Binding {
	return []key.Binding{k.Tab, k.AddItem, k.RemoveItem, k.Save, k.Cancel}
}

func (k keyMap) settingsShort() []key.Binding {
	return []key.Binding{k.Tab, k.Save, k.Cancel}
}
