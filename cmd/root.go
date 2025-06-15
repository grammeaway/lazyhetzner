package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"lazyhetzner/internal/message"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/config"
	"lazyhetzner/internal/model"
)




// Project list item
type projectItem struct {
	config    config.ProjectConfig
	isDefault bool
}

func (i projectItem) FilterValue() string { return i.config.Name }
func (i projectItem) Title() string {
	if i.isDefault {
		return fmt.Sprintf("⭐ %s", i.config.Name)
	}
	return i.config.Name
}
func (i projectItem) Description() string {
	tokenPreview := i.config.Token
	if len(tokenPreview) > 16 {
		tokenPreview = tokenPreview[:16] + "..."
	}
	if i.isDefault {
		return fmt.Sprintf("Token: %s (default project)", tokenPreview)
	}
	return fmt.Sprintf("Token: %s", tokenPreview)
}

// List items for different resources
type serverItem struct {
	server *hcloud.Server
}

func (i serverItem) FilterValue() string { return i.server.Name }
func (i serverItem) Title() string       { return i.server.Name }
func (i serverItem) Description() string {
	var statusDisplay string
	if i.server.Status == hcloud.ServerStatusRunning {
		statusDisplay = "🟢 " + string(i.server.Status)
	} else if i.server.Status == hcloud.ServerStatusStarting || i.server.Status == hcloud.ServerStatusInitializing {
		statusDisplay = "🟡 " + string(i.server.Status)
	} else {
		statusDisplay = "🔴 " + string(i.server.Status)
	}
	return fmt.Sprintf("%s | %s | %s", statusDisplay, i.server.ServerType.Name, i.server.PublicNet.IPv4.IP.String())
}

type networkItem struct {
	network *hcloud.Network
}

func (i networkItem) FilterValue() string { return i.network.Name }
func (i networkItem) Title() string       { return i.network.Name }
func (i networkItem) Description() string {
	return fmt.Sprintf("IP Range: %s | Subnets: %d", i.network.IPRange.String(), len(i.network.Subnets))
}

type loadBalancerItem struct {
	lb *hcloud.LoadBalancer
}

func (i loadBalancerItem) FilterValue() string { return i.lb.Name }
func (i loadBalancerItem) Title() string       { return i.lb.Name }
func (i loadBalancerItem) Description() string {
	status := "🟢 Available"
	if i.lb.PublicNet.Enabled {
		return fmt.Sprintf("%s | %s | Targets: %d", status, i.lb.PublicNet.IPv4.IP.String(), len(i.lb.Targets))
	}
	return fmt.Sprintf("%s | Private only | Targets: %d", status, len(i.lb.Targets))
}

type volumeItem struct {
	volume *hcloud.Volume
}

func (i volumeItem) FilterValue() string { return i.volume.Name }
func (i volumeItem) Title() string       { return i.volume.Name }
func (i volumeItem) Description() string {
	status := "📦 Available"
	if i.volume.Server != nil {
		status = "🔗 Attached to " + i.volume.Server.Name
	}
	return fmt.Sprintf("%s | %dGB | %s", status, i.volume.Size, i.volume.Location.Name)
}


// Messages

type resourcesLoadedMsg struct {
	servers       []*hcloud.Server
	networks      []*hcloud.Network
	loadBalancers []*hcloud.LoadBalancer
	volumes       []*hcloud.Volume
}


type serversLoadedMsg struct {
	servers []*hcloud.Server
}

type networksLoadedMsg struct {
	networks []*hcloud.Network
}

type loadBalancersLoadedMsg struct {
	loadBalancers []*hcloud.LoadBalancer
}

type volumesLoadedMsg struct {
	volumes []*hcloud.Volume
}

type resourceLoadStartMsg struct {
	resourceType model.ResourceType
}

// Commands

func loadResources(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		servers, err := client.Server.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		networks, err := client.Network.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		return resourcesLoadedMsg{
			servers:       servers,
			networks:      networks,
			loadBalancers: loadBalancers,
			volumes:       volumes,
		}
	}
}

func loadServers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		servers, err := client.Server.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return serversLoadedMsg{servers: servers}
	}
}

func loadNetworks(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		networks, err := client.Network.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return networksLoadedMsg{networks: networks}
	}
}

func loadLoadBalancers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return loadBalancersLoadedMsg{loadBalancers: loadBalancers}
	}
}

func loadVolumes(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return volumesLoadedMsg{volumes: volumes}
	}
}

func startResourceLoad(rt model.ResourceType) tea.Cmd {
	return func() tea.Msg {
		return resourceLoadStartMsg{resourceType: rt}
	}
}


func getResourceLabels(client *hcloud.Client, resourceType model.ResourceType, resourceID int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var labels map[string]string

		switch resourceType {
		case model.ResourceServers:
			server, _, err := client.Server.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if server == nil {
				return message.ErrorMsg{fmt.Errorf("server with ID %d not found", strconv.Itoa(resourceID))}
			}
			labels = server.Labels

		case model.ResourceNetworks:
			network, _, err := client.Network.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if network == nil {
				return message.ErrorMsg{fmt.Errorf("network with ID %d not found", resourceID)}
			}
			labels = network.Labels

		case model.ResourceLoadBalancers:
			lb, _, err := client.LoadBalancer.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if lb == nil {
				return message.ErrorMsg{fmt.Errorf("load balancer with ID %d not found", resourceID)}
			}
			labels = lb.Labels

		case model.ResourceVolumes:
			volume, _, err := client.Volume.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if volume == nil {
				return message.ErrorMsg{fmt.Errorf("volume with ID %d not found", resourceID)}
			}
			labels = volume.Labels

		default:
			return message.ErrorMsg{fmt.Errorf("unknown resource type: %d", resourceType)}
		}

		labelList := make([]string, 0, len(labels))
		for k, v := range labels {
			labelList = append(labelList, fmt.Sprintf("%s: %s", k, v))
		}

		return message.StatusMsg(fmt.Sprintf("Labels for resource ID %d:\n%s", resourceID, strings.Join(labelList, "\n")))
	}
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return message.ClipboardCopiedMsg(text)
	}
}


func clearStatusMessage() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return message.StatusMsg("")
	})
}


func initialModel() model.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter your Hetzner Cloud API token..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50

	lists := make(map[model.ResourceType]list.Model)
	loadedResources := make(map[model.ResourceType]bool)

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
