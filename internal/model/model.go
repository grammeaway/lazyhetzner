package model

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/config"
	ctm "github.com/grammeaway/lazyhetzner/internal/context_menu"
	ctm_fw "github.com/grammeaway/lazyhetzner/internal/context_menu/firewall"
	ctm_fip "github.com/grammeaway/lazyhetzner/internal/context_menu/floatingip"
	ctm_lb "github.com/grammeaway/lazyhetzner/internal/context_menu/loadbalancer"
	ctm_n "github.com/grammeaway/lazyhetzner/internal/context_menu/network"
	ctm_serv "github.com/grammeaway/lazyhetzner/internal/context_menu/server"
	ctm_vol "github.com/grammeaway/lazyhetzner/internal/context_menu/volume"
	"github.com/grammeaway/lazyhetzner/internal/input_form"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	r_fw "github.com/grammeaway/lazyhetzner/internal/resource/firewall"
	r_fip "github.com/grammeaway/lazyhetzner/internal/resource/floatingip"
	r_lb "github.com/grammeaway/lazyhetzner/internal/resource/loadbalancer"
	r_n "github.com/grammeaway/lazyhetzner/internal/resource/network"
	r_prj "github.com/grammeaway/lazyhetzner/internal/resource/project"
	r_serv "github.com/grammeaway/lazyhetzner/internal/resource/server"
	r_vol "github.com/grammeaway/lazyhetzner/internal/resource/volume"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"strconv"
	"time"
)

// Main model
type Model struct {
	State                      state
	config                     *config.Config
	TokenInput                 textinput.Model
	projectForm                input_form.InputForm
	projectList                list.Model
	client                     *hcloud.Client
	currentProject             string
	activeTab                  resource.ResourceType
	Lists                      map[resource.ResourceType]list.Model
	contextMenu                ctm.ContextMenu
	Help                       help.Model
	err                        error
	width                      int
	height                     int
	statusMessage              string
	LoadedResources            map[resource.ResourceType]bool
	loadingResource            resource.ResourceType
	IsLoading                  bool
	loadedLabels               map[string]string
	labelsPertainingToResource string
	loadbalancerBeingViewed    *hcloud.LoadBalancer
	loadbalancerTargets        []hcloud.LoadBalancerTarget
	loadbalancerServices       []hcloud.LoadBalancerService
	firewallBeingViewed        *hcloud.Firewall
	firewallRules              []hcloud.FirewallRule
}

// Resource types for tabs

var resourceTabs = []string{"Servers", "Networks", "Load Balancers", "Floating IPs", "Firewalls", "Volumes"}

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
	case resource.ResourceFloatingIPs:
		return r_fip.LoadFloatingIPs(m.client)
	case resource.ResourceFirewalls:
		return r_fw.LoadFirewalls(m.client)
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

func (m *Model) executeContextAction(selectedAction string, resourceType resource.ResourceType, resourceID int64) tea.Cmd {
	switch resourceType {
	case resource.ResourceServers:
		server, _, err := m.client.Server.Get(context.Background(), strconv.FormatInt(resourceID, 10))
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

		return ctm_serv.ExecuteServerContextAction(selectedAction, server, m.config.DefaultTerminal)
	case resource.ResourceNetworks:
		network, _, err := m.client.Network.Get(context.Background(), strconv.FormatInt(resourceID, 10))
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
	case resource.ResourceLoadBalancers:
		loadBalancer, _, err := m.client.LoadBalancer.Get(context.Background(), strconv.FormatInt(resourceID, 10))
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if loadBalancer == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("load balancer with ID %d not found", resourceID)}
			}
		}
		return ctm_lb.ExecuteLoadbalancerContextAction(selectedAction, loadBalancer)
	case resource.ResourceVolumes:
		volume, _, err := m.client.Volume.Get(context.Background(), strconv.FormatInt(resourceID, 10))
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if volume == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("volume with ID %d not found", resourceID)}
			}
		}
		// unfold server details if attached
		if volume.Server != nil {
			server, _, err := m.client.Server.GetByID(context.Background(), volume.Server.ID)
			if err == nil && server != nil {
				volume.Server = server
			}
		}
		return ctm_vol.ExecuteVolumeContextAction(selectedAction, volume)
	case resource.ResourceFirewalls:
		firewall, _, err := m.client.Firewall.GetByID(context.Background(), resourceID)
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if firewall == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("firewall with ID %d not found", resourceID)}
			}
		}
		return ctm_fw.ExecuteFirewallContextAction(selectedAction, firewall)
	case resource.ResourceFloatingIPs:
		floatingIP, _, err := m.client.FloatingIP.Get(context.Background(), strconv.FormatInt(resourceID, 10))
		if err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{err}
			}
		}
		if floatingIP == nil {
			return func() tea.Msg {
				return message.ErrorMsg{fmt.Errorf("floating IP with ID %d not found", resourceID)}
			}
		}
		if floatingIP.Server != nil {
			server, _, err := m.client.Server.GetByID(context.Background(), floatingIP.Server.ID)
			if err == nil && server != nil {
				floatingIP.Server = server
			}
		}
		return ctm_fip.ExecuteFloatingIPContextAction(selectedAction, floatingIP)
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
