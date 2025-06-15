package model

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
	"lazyhetzner/internal/config"
	ctm "lazyhetzner/internal/context_menu"
	ctm_serv "lazyhetzner/internal/context_menu/server"
	ctm_n "lazyhetzner/internal/context_menu/network"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/input_form"
	"lazyhetzner/internal/message"
	"lazyhetzner/internal/resource"
	r_serv "lazyhetzner/internal/resource/server"
	r_vol "lazyhetzner/internal/resource/volume"
	r_lb "lazyhetzner/internal/resource/loadbalancer"
	r_n"lazyhetzner/internal/resource/network"
	r_prj "lazyhetzner/internal/resource/project"
	"context"
	"time"
)

// Main model
type Model struct {
	State           state
	config          *config.Config
	TokenInput      textinput.Model
	projectForm     input_form.InputForm
	projectList     list.Model
	client          *hcloud.Client
	currentProject  string
	activeTab       resource.ResourceType
	Lists           map[resource.ResourceType]list.Model
	contextMenu     ctm.ContextMenu
	Help            help.Model
	err             error
	width           int
	height          int
	statusMessage   string
	LoadedResources map[resource.ResourceType]bool
	loadingResource resource.ResourceType
	IsLoading       bool
	loadedLabels map[string]string
	labelsPertainingToResource string
}

// Resource types for tabs

var resourceTabs = []string{"Servers", "Networks", "Load Balancers", "Volumes"}

func (m *Model) getResourceLoadCmd(rt resource.ResourceType) tea.Cmd {
	if m.client == nil {
		return nil
	}

	switch rt {
	case resource.ResourceServers:
		return r_serv.LoadServers(m.client)
	case resource.ResourceNetworks:
		return r_n.LoadNetworks(m.client)
	case resource.ResourceLoadBalancers:
		return r_lb.LoadLoadBalancers(m.client)
	case resource.ResourceVolumes:
		return r_vol.LoadVolumes(m.client)
	default:
		return nil
	}
}

func clearStatusMessage() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return message.StatusMsg("")
	})
}


func (m *Model) executeContextAction(selectedAction string, resourceType resource.ResourceType, resourceID int) tea.Cmd {	
	switch resourceType {
	case resource.ResourceServers:
		server, _, err := m.client.Server.Get(context.Background(), strconv.Itoa(resourceID))
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if server == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("server with ID %d not found", resourceID)}
			}
		}

		return ctm_serv.ExecuteServerContextAction(selectedAction, server)
	case resource.ResourceNetworks:
		network, _, err := m.client.Network.Get(context.Background(), strconv.Itoa(resourceID))
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if network == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("network with ID %d not found", resourceID)}
			}
		}
		return ctm_n.ExecuteNetworkContextAction(selectedAction, network)
	}

	return nil
}



func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, config.LoadConfigCmd())
}

func (m *Model) updateProjectList() {
	items := make([]list.Item, len(m.config.Projects))
	for i, project := range m.config.Projects {
		items[i] = r_prj.ProjectItem{
			Config:    project,
			IsDefault: project.Name == m.config.DefaultProject,
		}
	}

	m.projectList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
	m.projectList.Title = "Hetzner Cloud Projects"
}
