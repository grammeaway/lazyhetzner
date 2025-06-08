package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a1a"))

	selectedMenuStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#874BFD")).
				Foreground(lipgloss.Color("#FFFDF5")).
				Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

// App states
type state int

const (
	stateTokenInput state = iota
	stateLoading
	stateResourceView
	stateContextMenu
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

// Context menu items
type contextMenuItem struct {
	label  string
	action string
}

type contextMenu struct {
	items        []contextMenuItem
	selectedItem int
	server       *hcloud.Server
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
	contextMenu     contextMenu
	help            help.Model
	err             error
	width           int
	height          int
	statusMessage   string
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

type statusMsg string

type sshLaunchedMsg struct{}

type clipboardCopiedMsg string

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

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		if err != nil {
			return errorMsg{err}
		}
		return clipboardCopiedMsg(text)
	}
}

func launchSSH(ip string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		
		// Determine terminal based on OS
		switch runtime.GOOS {
		case "darwin": // macOS
			cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "ssh root@%s"`, ip))
		case "linux":
			// Try common terminal emulators
			terminals := []string{"gnome-terminal", "konsole", "xterm", "alacritty", "kitty"}
			for _, term := range terminals {
				if _, err := exec.LookPath(term); err == nil {
					switch term {
					case "gnome-terminal":
						cmd = exec.Command(term, "--", "ssh", fmt.Sprintf("root@%s", ip))
					case "konsole":
						cmd = exec.Command(term, "-e", "ssh", fmt.Sprintf("root@%s", ip))
					default:
						cmd = exec.Command(term, "-e", "ssh", fmt.Sprintf("root@%s", ip))
					}
					break
				}
			}
		case "windows":
			// Use Windows Terminal if available, otherwise cmd
			if _, err := exec.LookPath("wt"); err == nil {
				cmd = exec.Command("wt", "ssh", fmt.Sprintf("root@%s", ip))
			} else {
				cmd = exec.Command("cmd", "/c", "start", "ssh", fmt.Sprintf("root@%s", ip))
			}
		}
		
		if cmd == nil {
			return errorMsg{fmt.Errorf("no suitable terminal found")}
		}
		
		err := cmd.Start()
		if err != nil {
			return errorMsg{err}
		}
		
		return sshLaunchedMsg{}
	}
}

func clearStatusMessage() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return statusMsg("")
	})
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
			case key.Matches(msg, keys.Enter):
				// Show context menu for servers
				if m.activeTab == resourceServers {
					if currentList, exists := m.lists[resourceServers]; exists {
						if selectedItem := currentList.SelectedItem(); selectedItem != nil {
							if serverItem, ok := selectedItem.(serverItem); ok {
								m.contextMenu = contextMenu{
									items: []contextMenuItem{
										{label: "📋 Copy Public IP", action: "copy_public_ip"},
										{label: "📋 Copy Private IP", action: "copy_private_ip"},
										{label: "🔗 SSH (New Terminal)", action: "ssh_new_terminal"},
										{label: "🔗 SSH (Current Terminal)", action: "ssh_current_terminal"},
										{label: "❌ Cancel", action: "cancel"},
									},
									selectedItem: 0,
									server:       serverItem.server,
								}
								m.state = stateContextMenu
							}
						}
					}
				}
				
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
			
		case stateContextMenu:
			switch {
			case key.Matches(msg, keys.Up):
				if m.contextMenu.selectedItem > 0 {
					m.contextMenu.selectedItem--
				}
				
			case key.Matches(msg, keys.Down):
				if m.contextMenu.selectedItem < len(m.contextMenu.items)-1 {
					m.contextMenu.selectedItem++
				}
				
			case key.Matches(msg, keys.Enter):
				selectedAction := m.contextMenu.items[m.contextMenu.selectedItem].action
				server := m.contextMenu.server
				m.state = stateResourceView
				
				switch selectedAction {
				case "copy_public_ip":
					if server.PublicNet.IPv4.IP != nil {
						return m, copyToClipboard(server.PublicNet.IPv4.IP.String())
					}
				case "copy_private_ip":
					if len(server.PrivateNet) > 0 && server.PrivateNet[0].IP != nil {
						return m, copyToClipboard(server.PrivateNet[0].IP.String())
					}
				case "ssh_new_terminal":
					if server.PublicNet.IPv4.IP != nil {
						return m, launchSSH(server.PublicNet.IPv4.IP.String())
					}
				case "ssh_current_terminal":
					if server.PublicNet.IPv4.IP != nil {
						// Suspend TUI and run SSH in current terminal
						return m, tea.ExecProcess(exec.Command("ssh", fmt.Sprintf("root@%s", server.PublicNet.IPv4.IP.String())), func(err error) tea.Msg {
							if err != nil {
								return errorMsg{err}
							}
							return sshLaunchedMsg{}
						})
					}
				}
				
			case key.Matches(msg, keys.Quit):
				m.state = stateResourceView
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
		
	case clipboardCopiedMsg:
		m.statusMessage = fmt.Sprintf("✅ Copied %s to clipboard", string(msg))
		return m, clearStatusMessage()
		
	case sshLaunchedMsg:
		m.statusMessage = "🚀 SSH session launched"
		return m, clearStatusMessage()
		
	case statusMsg:
		m.statusMessage = string(msg)
		
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
		
		// Status message
		statusView := ""
		if m.statusMessage != "" {
			statusView = "\n" + successStyle.Render(m.statusMessage)
		}
		
		helpText := "Tab: switch view • ←/→: navigate tabs • Enter: actions • q: quit"
		if m.activeTab == resourceServers {
			helpText = "Tab: switch view • ←/→: navigate tabs • Enter: server actions • q: quit"
		}
		
		return fmt.Sprintf(
			"%s\n\n%s%s\n\n%s",
			tabsView,
			listView,
			statusView,
			helpStyle.Render(helpText),
		)
		
	case stateContextMenu:
		// Render the current resource view in background
		var tabs []string
		for i, tab := range resourceTabs {
			if resourceType(i) == m.activeTab {
				tabs = append(tabs, titleStyle.Render(tab))
			} else {
				tabs = append(tabs, helpStyle.Render(tab))
			}
		}
		tabsView := strings.Join(tabs, " ")
		
		var listView string
		if currentList, exists := m.lists[m.activeTab]; exists {
			listView = currentList.View()
		}
		
		// Render context menu
		var menuItems []string
		for i, item := range m.contextMenu.items {
			if i == m.contextMenu.selectedItem {
				menuItems = append(menuItems, selectedMenuStyle.Render(item.label))
			} else {
				menuItems = append(menuItems, item.label)
			}
		}
		
		menuContent := strings.Join(menuItems, "\n")
		menu := menuStyle.Render(fmt.Sprintf("Actions for %s:\n\n%s", m.contextMenu.server.Name, menuContent))
		
		// Center the menu
		menuHeight := strings.Count(menu, "\n") + 1
		menuWidth := 40 // Approximate width
		_ = menuHeight  // Suppress unused variable warning
		_ = menuWidth   // Suppress unused variable warning
		
		// Position menu overlay
		menuOverlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menu)
		
		background := fmt.Sprintf("%s\n\n%s", tabsView, listView)
		
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, background) + menuOverlay
		
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
