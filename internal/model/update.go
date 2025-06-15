package model 

import (
	"fmt"
	"strings"
	"lazyhetzner/internal/config"
	ctm "lazyhetzner/internal/context_menu"
	ctm_serv "lazyhetzner/internal/context_menu/server"
	"lazyhetzner/internal/input_form"
	"lazyhetzner/internal/input_form/project"
	"lazyhetzner/internal/message"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/key"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"context"
)


func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update list sizes
		for rt := range m.Lists {
			l := m.Lists[rt]
			l.SetWidth(msg.Width - 4)
			l.SetHeight(msg.Height - 10)
			m.Lists[rt] = l
		}

		if m.config != nil {
			m.updateProjectList()
		}

		// Update form inputs
		for i := range m.projectForm.Inputs {
			m.projectForm.Inputs[i].Width = min(50, msg.Width-10)
		}

	case configLoadedMsg:
		m.config = msg.config
		m.updateProjectList()

		// If there's a default project, set it up but don't load resources yet
		if m.config.DefaultProject != "" {
			project := m.config.GetProject(m.config.DefaultProject)
			if project != nil {
				m.client = hcloud.NewClient(hcloud.WithToken(project.Token))
				m.currentProject = project.Name
				m.State = stateResourceView
				// Load the first tab's resources
				return m, tea.Batch(
					startResourceLoad(m.activeTab),
					m.getResourceLoadCmd(m.activeTab),
				)
			}
		}

		// If no projects, go to token input
		if len(m.config.Projects) == 0 {
			m.State = stateTokenInput
		}
	case tea.KeyMsg:
		switch m.State {
		case stateProjectSelect:
			switch {
			case key.Matches(msg, keys.Enter):
				if selectedItem := m.projectList.SelectedItem(); selectedItem != nil {
					if projectItem, ok := selectedItem.(projectItem); ok {
						m.client = hcloud.NewClient(hcloud.WithToken(projectItem.config.Token))
						m.currentProject = projectItem.config.Name
						m.state = stateResourceView
						// Reset loaded resources for new project
						m.loadedResources = make(map[ResourceType]bool)
						// Load the first tab's resources
						return m, tea.Batch(
							startResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			case key.Matches(msg, keys.Add):
				m.projectForm = project.NewProjectForm()
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
				m.projectForm.FocusIdx = (m.projectForm.FocusIdx + 1) % len(m.projectForm.Inputs)
				for i := range m.projectForm.Inputs {
					if i == m.projectForm.FocusIdx {
						m.projectForm.Inputs[i].Focus()
					} else {
						m.projectForm.Inputs[i].Blur()
					}
				}

			case key.Matches(msg, keys.Enter):
				// Submit form
				name := strings.TrimSpace(m.projectForm.Inputs[0].Value())
				token := strings.TrimSpace(m.projectForm.Inputs[1].Value())

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
				m.state = stateResourceView
				// Reset loaded resources
				m.loadedResources = make(map[ResourceType]bool)
				// Load the first tab's resources
				return m, tea.Batch(
					startResourceLoad(m.activeTab),
					m.getResourceLoadCmd(m.activeTab),
				)
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
								m.contextMenu = ctm_serv.CreateServerContextMenu(serverItem.server) 
								m.state = stateContextMenu
							}
						}
					}
				}

			case key.Matches(msg, keys.Tab):
				m.activeTab = (m.activeTab + 1) % resourceType(len(resourceTabs))

				// Load resources for new tab if not already loaded
				if !m.loadedResources[m.activeTab] {
					return m, tea.Batch(
						startResourceLoad(m.activeTab),
						m.getResourceLoadCmd(m.activeTab),
					)
				}
			case key.Matches(msg, keys.Reload):
				if m.client != nil {
					// Mark current resource as not loaded and reload
					m.loadedResources[m.activeTab] = false
					return m, tea.Batch(
						startResourceLoad(m.activeTab),
						m.getResourceLoadCmd(m.activeTab),
					)
				}
			case key.Matches(msg, keys.Left):
				if m.activeTab > 0 {
					m.activeTab--

					// Load resources for new tab if not already loaded
					if !m.loadedResources[m.activeTab] {
						return m, tea.Batch(
							startResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			case key.Matches(msg, keys.Right):
				if int(m.activeTab) < len(resourceTabs)-1 {
					m.activeTab++

					// Load resources for new tab if not already loaded
					if !m.loadedResources[m.activeTab] {
						return m, tea.Batch(
							startResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			case key.Matches(msg, keys.Quit):
				m.state = stateProjectSelect
			}

		case stateContextMenu:
			switch {
			case key.Matches(msg, keys.Up):
				if m.contextMenu.SelectedItem > 0 {
					m.contextMenu.SelectedItem--
				}

			case key.Matches(msg, keys.Down):
				if m.contextMenu.SelectedItem < len(m.contextMenu.Items)-1 {
					m.contextMenu.SelectedItem++
				}

			case key.Matches(msg, keys.Enter):
				selectedAction := m.contextMenu.Items[m.contextMenu.SelectedItem].Action
				server := m.contextMenu.Server
				m.state = stateResourceView
				return m, m.executeContextAction(selectedAction, server)

			case key.Matches(msg, keys.Num1, keys.Num2, keys.Num3, keys.Num4, keys.Num5,
				keys.Num6, keys.Num7, keys.Num8, keys.Num9, keys.Num0):

				keyStr := msg.String()
				selectedIndex := getIndexFromNumber(keyStr)

				// Check if the number corresponds to a valid menu item
				if selectedIndex >= 0 && selectedIndex < len(m.contextMenu.Items) {
					selectedAction := m.contextMenu.Items[selectedIndex].Action
					server := m.contextMenu.Server
					m.state = stateResourceView
					return m, m.executeContextAction(selectedAction, server)
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

	case resourceLoadStartMsg:
		m.isLoading = true
		m.loadingResource = msg.resourceType

	case serversLoadedMsg:
		m.isLoading = false
		m.loadedResources[resourceServers] = true

		// Create server list
		serverItems := make([]list.Item, len(msg.servers))
		for i, server := range msg.servers {
			serverItems[i] = serverItem{server: server}
		}

		serversList := list.New(serverItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		serversList.Title = "Servers"
		m.lists[resourceServers] = serversList

	case networksLoadedMsg:
		m.isLoading = false
		m.loadedResources[resourceNetworks] = true

		// Create network list
		networkItems := make([]list.Item, len(msg.networks))
		for i, network := range msg.networks {
			networkItems[i] = networkItem{network: network}
		}

		networksList := list.New(networkItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		networksList.Title = "Networks"
		m.lists[resourceNetworks] = networksList

	case loadBalancersLoadedMsg:
		m.isLoading = false
		m.loadedResources[resourceLoadBalancers] = true

		// Create load balancer list
		lbItems := make([]list.Item, len(msg.loadBalancers))
		for i, lb := range msg.loadBalancers {
			lbItems[i] = loadBalancerItem{lb: lb}
		}

		lbList := list.New(lbItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		lbList.Title = "Load Balancers"
		m.lists[resourceLoadBalancers] = lbList

	case volumesLoadedMsg:
		m.isLoading = false
		m.loadedResources[resourceVolumes] = true

		// Create volume list
		volumeItems := make([]list.Item, len(msg.volumes))
		for i, volume := range msg.volumes {
			volumeItems[i] = volumeItem{volume: volume}
		}

		volumesList := list.New(volumeItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		volumesList.Title = "Volumes"
		m.lists[resourceVolumes] = volumesList

	case message.ClipboardCopiedMsg:
		m.statusMessage = fmt.Sprintf("✅ Copied %s to clipboard", string(msg))
		return m, clearStatusMessage()

	case sshLaunchedMsg:
		m.statusMessage = "🚀 SSH session launched"
		return m, clearStatusMessage()
	case tmuxSSHLaunchedMsg:
		m.statusMessage = "🪟 SSH session launched in tmux"
		return m, clearStatusMessage()

	case zellijSSHLaunchedMsg:
		m.statusMessage = "📱 SSH session launched in zellij"
		return m, clearStatusMessage()

	case message.StatusMsg:
		m.statusMessage = string(msg)

	case message.ErrorMsg:
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
