package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/model"
	"github.com/grammeaway/lazyhetzner/internal/resource"
)

var (
	version = "nightly"
	commit  = "unknown"
	date    = "unknown"
)

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
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("awsbreeze version: %s\ncommit: %s\nbuilt at: %s\n", version, commit, date)
		return
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
