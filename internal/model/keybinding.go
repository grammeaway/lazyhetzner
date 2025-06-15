package model

import (
	"github.com/charmbracelet/bubbles/key"
)

// Key bindings
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Tab    key.Binding
	Enter  key.Binding
	Delete key.Binding
	Add    key.Binding
	Quit   key.Binding
	Help   key.Binding
	Reload key.Binding

	Num1 key.Binding
	Num2 key.Binding
	Num3 key.Binding
	Num4 key.Binding
	Num5 key.Binding
	Num6 key.Binding
	Num7 key.Binding
	Num8 key.Binding
	Num9 key.Binding
	Num0 key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Enter, k.Add, k.Delete},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch view"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "delete"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add project"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload resources"),
	),

	Num1: key.NewBinding(key.WithKeys("1")),
	Num2: key.NewBinding(key.WithKeys("2")),
	Num3: key.NewBinding(key.WithKeys("3")),
	Num4: key.NewBinding(key.WithKeys("4")),
	Num5: key.NewBinding(key.WithKeys("5")),
	Num6: key.NewBinding(key.WithKeys("6")),
	Num7: key.NewBinding(key.WithKeys("7")),
	Num8: key.NewBinding(key.WithKeys("8")),
	Num9: key.NewBinding(key.WithKeys("9")),
	Num0: key.NewBinding(key.WithKeys("0")),
}

