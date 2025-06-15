package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"lazyhetzner/internal/config"
	ctm "lazyhetzner/internal/context_menu"
	ctm_serv "lazyhetzner/internal/context_menu/server"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/hetznercloud/hcloud-go/hcloud"

	"lazyhetzner/internal/input_form"
	"lazyhetzner/internal/message"
	"context"
)

// Main model
type Model struct {
	state           state
	config          *config.Config
	tokenInput      textinput.Model
	projectForm     input_form.InputForm
	projectList     list.Model
	client          *hcloud.Client
	currentProject  string
	activeTab       ResourceType
	lists           map[ResourceType]list.Model
	contextMenu     ctm.ContextMenu
	help            help.Model
	err             error
	width           int
	height          int
	statusMessage   string
	loadedResources map[ResourceType]bool
	loadingResource ResourceType
	isLoading       bool
}

// Resource types for tabs
type ResourceType int

const (
	ResourceServers ResourceType = iota
	ResourceNetworks
	ResourceLoadBalancers
	ResourceVolumes
)

func (m *Model) getResourceLoadCmd(rt ResourceType) tea.Cmd {
	if m.client == nil {
		return nil
	}

	switch rt {
	case ResourceServers:
		return loadServers(m.client)
	case ResourceNetworks:
		return loadNetworks(m.client)
	case ResourceLoadBalancers:
		return loadLoadBalancers(m.client)
	case ResourceVolumes:
		return loadVolumes(m.client)
	default:
		return nil
	}
}




func (m *Model) executeContextAction(selectedAction string, resourceType ResourceType, resourceID int) tea.Cmd {	
	switch resourceType {
	case ResourceServers:
		server, _, err := m.client.Server.Get(context.Background(), strconv.Itoa(resourceID))
		
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}

		return ctm_serv.ExecuteServerContextAction(selectedAction, server)
	}
	return nil
}



func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, loadConfigCmd())
}

func (m *Model) updateProjectList() {
	items := make([]list.Item, len(m.config.Projects))
	for i, project := range m.config.Projects {
		items[i] = projectItem{
			config:    project,
			isDefault: project.Name == m.config.DefaultProject,
		}
	}

	m.projectList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
	m.projectList.Title = "Hetzner Cloud Projects"
}
