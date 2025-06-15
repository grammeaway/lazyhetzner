package cmd

import (
	"fmt"
	"os"

	"lazyhetzner/internal/message"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"lazyhetzner/internal/model"
	"lazyhetzner/internal/resource"
)


func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return message.ClipboardCopiedMsg(text)
	}
}


func initialModel() model.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter your Hetzner Cloud API token..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50

	lists := make(map[resource.ResourceType]list.Model)
	loadedResources := make(map[resource.ResourceType]bool)

	return model.Model{
		State:           model.StateProjectSelect,
		TokenInput:      ti,
		Lists:           lists,
		Help:            help.New(),
		LoadedResources: loadedResources,
		IsLoading:       false,
	}
}


func App() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
