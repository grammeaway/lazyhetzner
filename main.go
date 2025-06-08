package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// Config management
type ProjectConfig struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type Config struct {
	Projects       []ProjectConfig `json:"projects"`
	DefaultProject string          `json:"default_project"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".config", "lazyhetzner")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	
	return filepath.Join(configDir, "config.json"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{Projects: []ProjectConfig{}}, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Projects: []ProjectConfig{}}, nil
		}
		return nil, err
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) AddProject(name, token string) {
	// Remove existing project with same name
	for i, p := range c.Projects {
		if p.Name == name {
			c.Projects = append(c.Projects[:i], c.Projects[i+1:]...)
			break
		}
	}
	
	c.Projects = append(c.Projects, ProjectConfig{
		Name:  name,
		Token: token,
	})
	
	// Set as default if it's the first project
	if len(c.Projects) == 1 {
		c.DefaultProject = name
	}
}

func (c *Config) GetProject(name string) *ProjectConfig {
	for _, p := range c.Projects {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func (c *Config) RemoveProject(name string) {
	for i, p := range c.Projects {
		if p.Name == name {
			c.Projects = append(c.Projects[:i], c.Projects[i+1:]...)
			if c.DefaultProject == name && len(c.Projects) > 0 {
				c.DefaultProject = c.Projects[0].Name
			}
			break
		}
	}
}

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

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	focusedStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#874BFD"))

	blurredStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#626262"))
)

// App states
type state int

const (
	stateProjectSelect state = iota
	stateProjectManage
	stateTokenInput
	stateLoading
	stateResourceView
	stateContextMenu
	stateError
)

// Input forms
type inputForm struct {
	inputs    []textinput.Model
	focusIdx  int
	submitBtn string
	cancelBtn string
}

func newProjectForm() inputForm {
	inputs := make([]textinput.Model, 2)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Project name (e.g., production, staging)"
	inputs[0].Focus()
	inputs[0].Width = 40
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Hetzner Cloud API token"
	inputs[1].Width = 50
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	
	return inputForm{
		inputs:    inputs,
		focusIdx:  0,
		submitBtn: "Add Project",
		cancelBtn: "Cancel",
	}
}

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

// Project list item
type projectItem struct {
	config    ProjectConfig
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
	state           state
	config          *Config
	tokenInput      textinput.Model
	projectForm     inputForm
	projectList     list.Model
	client          *hcloud.Client
	currentProject  string
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
type configLoadedMsg struct {
	config *Config
}

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

type projectSavedMsg struct{}

// Commands
func loadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		config, err := loadConfig()
		if err != nil {
			return errorMsg{err}
		}
		return configLoadedMsg{config}
	}
}

func saveConfigCmd(config *Config) tea.Cmd {
	return func() tea.Msg {
		if err := saveConfig(config); err != nil {
			return errorMsg{err}
		}
		return projectSavedMsg{}
	}
}

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
			terminals := []string{"gnome-terminal", "konsole", "xterm", "alacritty", "kitty", "foot"}
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
		state:      stateProjectSelect,
		tokenInput: ti,
		lists:      lists,
		help:       help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, loadConfigCmd())
}

func (m *model) updateProjectList() {
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
		
		if m.config != nil {
			m.updateProjectList()
		}
		
		// Update form inputs
		for i := range m.projectForm.inputs {
			m.projectForm.inputs[i].Width = min(50, msg.Width-10)
		}
		
	case configLoadedMsg:
		m.config = msg.config
		m.updateProjectList()
		
		// If there's a default project, load it automatically
		if m.config.DefaultProject != "" {
			project := m.config.GetProject(m.config.DefaultProject)
			if project != nil {
				m.client = hcloud.NewClient(hcloud.WithToken(project.Token))
				m.currentProject = project.Name
				m.state = stateLoading
				return m, loadResources(m.client)
			}
		}
		
		// If no projects, go to token input
		if len(m.config.Projects) == 0 {
			m.state = stateTokenInput
		}
		
	case tea.KeyMsg:
		switch m.state {
		case stateProjectSelect:
			switch {
			case key.Matches(msg, keys.Enter):
				if selectedItem := m.projectList.SelectedItem(); selectedItem != nil {
					if projectItem, ok := selectedItem.(projectItem); ok {
						m.client = hcloud.NewClient(hcloud.WithToken(projectItem.config.Token))
						m.currentProject = projectItem.config.Name
						m.state = stateLoading
						return m, loadResources(m.client)
					}
				}
				
			case key.Matches(msg, keys.Add):
				m.projectForm = newProjectForm()
				m.state = stateProjectManage
				
			case key.Matches(msg, keys.Delete):
				if selectedItem := m.projectList.SelectedItem(); selectedItem != nil {
					if projectItem, ok := selectedItem.(projectItem); ok {
						m.config.RemoveProject(projectItem.config.Name)
						m.updateProjectList()
						return m, saveConfigCmd(m.config)
					}
				}
				
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			}
			
		case stateProjectManage:
			switch {
			case key.Matches(msg, keys.Tab):
				m.projectForm.focusIdx = (m.projectForm.focusIdx + 1) % len(m.projectForm.inputs)
				for i := range m.projectForm.inputs {
					if i == m.projectForm.focusIdx {
						m.projectForm.inputs[i].Focus()
					} else {
						m.projectForm.inputs[i].Blur()
					}
				}
				
			case key.Matches(msg, keys.Enter):
				// Submit form
				name := strings.TrimSpace(m.projectForm.inputs[0].Value())
				token := strings.TrimSpace(m.projectForm.inputs[1].Value())
				
				if name != "" && token != "" {
					m.config.AddProject(name, token)
					m.updateProjectList()
					m.state = stateProjectSelect
					return m, saveConfigCmd(m.config)
				}
				
			case key.Matches(msg, keys.Quit):
				m.state = stateProjectSelect
			}
			
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
				if len(m.config.Projects) > 0 {
					m.state = stateProjectSelect
				} else {
					return m, tea.Quit
				}
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
				m.state = stateProjectSelect
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
		
	case projectSavedMsg:
		m.statusMessage = "✅ Project configuration saved"
		return m, clearStatusMessage()
		
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
	
	if m.state == stateProjectSelect && m.config != nil {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		return m, cmd
	}
	
	if m.state == stateProjectManage {
		var cmd tea.Cmd
		m.projectForm.inputs[m.projectForm.focusIdx], cmd = m.projectForm.inputs[m.projectForm.focusIdx].Update(msg)
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
	case stateProjectSelect:
		if m.config == nil {
			return fmt.Sprintf(
				"\n%s\n\n%s\n",
				titleStyle.Render("Hetzner Cloud TUI"),
				infoStyle.Render("Loading configuration..."),
			)
		}
		
		if len(m.config.Projects) == 0 {
			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s\n",
				titleStyle.Render("Hetzner Cloud TUI"),
				warningStyle.Render("No projects configured yet."),
				helpStyle.Render("Press 'a' to add your first project • Press q to quit"),
			)
		}
		
		// Status message
		statusView := ""
		if m.statusMessage != "" {
			statusView = "\n" + successStyle.Render(m.statusMessage)
		}
		
		return fmt.Sprintf(
			"\n%s\n\n%s%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI"),
			m.projectList.View(),
			statusView,
			helpStyle.Render("Enter: select project • a: add project • d: delete project • q: quit"),
		)
		
	case stateProjectManage:
		var formView strings.Builder
		formView.WriteString("Add New Project\n\n")
		
		for i, input := range m.projectForm.inputs {
			var style lipgloss.Style
			if i == m.projectForm.focusIdx {
				style = focusedStyle
			} else {
				style = blurredStyle
			}
			
			label := "Project Name:"
			if i == 1 {
				label = "API Token:"
			}
			
			formView.WriteString(fmt.Sprintf("%s\n%s\n\n", label, style.Render(input.View())))
		}
		
		formView.WriteString(helpStyle.Render("Tab: next field • Enter: save • Esc: cancel"))
		
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI - Add Project"),
			formView.String(),
		)
		
	case stateTokenInput:
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI"),
			infoStyle.Render("Enter API token for one-time access:"),
			m.tokenInput.View(),
			helpStyle.Render("Press Enter to continue • Press q to go back"),
		)
		
	case stateLoading:
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			titleStyle.Render("Hetzner Cloud TUI"),
			infoStyle.Render("Loading resources..."),
		)
		
	case stateResourceView:
		// Project header
		projectHeader := fmt.Sprintf("Project: %s", m.currentProject)
		if m.currentProject == "" {
			projectHeader = "One-time Access"
		}
		
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
		
		helpText := "Tab: switch view • ←/→: navigate tabs • Enter: actions • q: back to projects"
		if m.activeTab == resourceServers {
			helpText = "Tab: switch view • ←/→: navigate tabs • Enter: server actions • q: back to projects"
		}
		
		return fmt.Sprintf(
			"%s\n%s\n\n%s%s\n\n%s",
			infoStyle.Render(projectHeader),
			tabsView,
			listView,
			statusView,
			helpStyle.Render(helpText),
		)
		
	case stateContextMenu:
		// Render the current resource view in background
		projectHeader := fmt.Sprintf("Project: %s", m.currentProject)
		if m.currentProject == "" {
			projectHeader = "One-time Access"
		}
		
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
		
		background := fmt.Sprintf("%s\n%s\n\n%s", infoStyle.Render(projectHeader), tabsView, listView)
		
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

// Helper function for min (Go 1.21+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
