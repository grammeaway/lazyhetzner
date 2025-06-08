package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))
)

// App states
type state int

const (
	stateTokenInput state = iota
	stateLoading
	stateResourceView
	stateError
)

// Key bindings
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Tab    key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Help   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Enter},
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
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Resource types for tabs
type resourceType int

const (
	resourceServers resourceType = iota
	resourceNetworks
	resourceLoadBalancers
	resourceVolumes
)

var resourceTabs = []string{"Servers", "Networks", "Load Balancers", "Volumes"}

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

// Main model
type model struct {
	state            state
	tokenInput       textinput.Model
	client          *hcloud.Client
	activeTab       resourceType
	lists           map[resourceType]list.Model
	help            help.Model
	err             error
	width           int
	height          int
}

// Messages
type resourcesLoadedMsg struct {
	servers       []*hcloud.Server
	networks      []*hcloud.Network
	loadBalancers []*hcloud.LoadBalancer
	volumes       []*hcloud.Volume
}

type errorMsg struct {
	err error
}

func (e errorMsg) Error() string { return e.err.Error() }

// Commands
func loadResources(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		
		servers, err := client.Server.All(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		networks, err := client.Network.All(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		return resourcesLoadedMsg{
			servers:       servers,
			networks:      networks,
			loadBalancers: loadBalancers,
			volumes:       volumes,
		}
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your Hetzner Cloud API token..."
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50

	lists := make(map[resourceType]list.Model)
	
	return model{
		state:      stateTokenInput,
		tokenInput: ti,
		lists:      lists,
		help:       help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update list sizes
		for rt := range m.lists {
			l := m.lists[rt]
			l.SetWidth(msg.Width - 4)
			l.SetHeight(msg.Height - 10)
			m.lists[rt] = l
		}
		
	case tea.KeyMsg:
		switch m.state {
		case stateTokenInput:
			switch {
			case key.Matches(msg, keys.Enter):
				token := strings.TrimSpace(m.tokenInput.Value())
				if token == "" {
					return m, nil
				}
				
				m.client = hcloud.NewClient(hcloud.WithToken(token))
				m.state = stateLoading
				return m, loadResources(m.client)
				
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			}
			
		case stateResourceView:
			switch {
			case key.Matches(msg, keys.Tab):
				m.activeTab = (m.activeTab + 1) % resourceType(len(resourceTabs))
				
			case key.Matches(msg, keys.Left):
				if m.activeTab > 0 {
					m.activeTab--
				}
				
			case key.Matches(msg, keys.Right):
				if int(m.activeTab) < len(resourceTabs)-1 {
					m.activeTab++
				}
				
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			}
			
		case stateError:
			if key.Matches(msg, keys.Quit) {
				return m, tea.Quit
			}
		}
		
	case resourcesLoadedMsg:
		m.state = stateResourceView
		
		// Create lists for each resource type
		serverItems := make([]list.Item, len(msg.servers))
		for i, server := range msg.servers {
			serverItems[i] = serverItem{server: server}
		}
		
		networkItems := make([]list.Item, len(msg.networks))
		for i, network := range msg.networks {
			networkItems[i] = networkItem{network: network}
		}
		
		lbItems := make([]list.Item, len(msg.loadBalancers))
		for i, lb := range msg.loadBalancers {
			lbItems[i] = loadBalancerItem{lb: lb}
		}
		
		volumeItems := make([]list.Item, len(msg.volumes))
		for i, volume := range msg.volumes {
			volumeItems[i] = volumeItem{volume: volume}
		}
		
		// Initialize lists
		serversList := list.New(serverItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		serversList.Title = "Servers"
		m.lists[resourceServers] = serversList
		
		networksList := list.New(networkItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		networksList.Title = "Networks"
		m.lists[resourceNetworks] = networksList
		
		lbList := list.New(lbItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		lbList.Title = "Load Balancers"
		m.lists[resourceLoadBalancers] = lbList
		
		volumesList := list.New(volumeItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		volumesList.Title = "Volumes"
		m.lists[resourceVolumes] = volumesList
		
	case errorMsg:
		m.state = stateError
		m.err = msg.err
	}
	
	// Update components
	if m.state == stateTokenInput {
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	}
	
	if m.state == stateResourceView {
		if currentList, exists := m.lists[m.activeTab]; exists {
			var cmd tea.Cmd
			m.lists[m.activeTab], cmd = currentList.Update(msg)
			return m, cmd
		}
	}
	
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateTokenInput:
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI"),
			m.tokenInput.View(),
			helpStyle.Render("Press Enter to continue • Press q to quit"),
		)
		
	case stateLoading:
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI"),
			infoStyle.Render("Loading resources..."),
		)
		
	case stateResourceView:
		// Render tabs
		var tabs []string
		for i, tab := range resourceTabs {
			if resourceType(i) == m.activeTab {
				tabs = append(tabs, titleStyle.Render(tab))
			} else {
				tabs = append(tabs, helpStyle.Render(tab))
			}
		}
		tabsView := strings.Join(tabs, " ")
		
		// Render current list
		var listView string
		if currentList, exists := m.lists[m.activeTab]; exists {
			listView = currentList.View()
		}
		
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			tabsView,
			listView,
			helpStyle.Render("Tab: switch view • ←/→: navigate tabs • q: quit"),
		)
		
	case stateError:
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI - Error"),
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
			helpStyle.Render("Press q to quit"),
		)
	}
	
	return ""
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
