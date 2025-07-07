
package model 

import (
	"fmt"
	"strings"
	"github.com/grammeaway/lazyhetzner/internal/config"
	ctm_serv "github.com/grammeaway/lazyhetzner/internal/context_menu/server"
	ctm_n "github.com/grammeaway/lazyhetzner/internal/context_menu/network"
	ctm_lb "github.com/grammeaway/lazyhetzner/internal/context_menu/loadbalancer"
	"github.com/grammeaway/lazyhetzner/internal/input_form/project"
	"github.com/grammeaway/lazyhetzner/internal/message"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/key"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	util "github.com/grammeaway/lazyhetzner/utility"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	r_prj "github.com/grammeaway/lazyhetzner/internal/resource/project"
	r_serv "github.com/grammeaway/lazyhetzner/internal/resource/server"
	r_vol "github.com/grammeaway/lazyhetzner/internal/resource/volume"
	r_lb "github.com/grammeaway/lazyhetzner/internal/resource/loadbalancer"
	r_n "github.com/grammeaway/lazyhetzner/internal/resource/network"
	r_label "github.com/grammeaway/lazyhetzner/internal/resource/label"
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

		// Update form.Inputs
		for i := range m.projectForm.Inputs {
			m.projectForm.Inputs[i].Width = min(50, msg.Width-10)
		}

	case config.ConfigLoadedMsg:
		m.config = msg.Config
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
					resource.StartResourceLoad(m.activeTab),
					m.getResourceLoadCmd(m.activeTab),
				)
			}
		}

		// If no projects, go to token input
		if len(m.config.Projects) == 0 {
			m.State = stateTokenInput
		}
	case tea.KeyMsg:
		if key.Matches(msg, keys.Exit) {
			// Handle exit key - quit the application
			return m, tea.Quit
		}
		// Handle global quit first - only quit the entire app from specific states
		if key.Matches(msg, keys.Quit) {
			switch m.State {
			case StateProjectSelect:
				// Only quit the entire application from project select
				return m, tea.Quit
			case stateTokenInput:
				// From token input, go back to project select if projects exist, otherwise quit
				if len(m.config.Projects) > 0 {
					m.State = StateProjectSelect
					return m, nil
				} else {
					return m, tea.Quit
				}
			case stateError:
				// Quit from error state
				return m, tea.Quit
			case stateLabelView:
				// From label view, go back to resource view
				m.State = stateResourceView
				return m, nil
			case stateResourceView:
				// From resource view, go back to project select
				m.State = StateProjectSelect
				return m, nil
			case stateProjectManage:
				// From project manage, go back to project select
				m.State = StateProjectSelect
				return m, nil
			case stateContextMenu:
				// From context menu, go back to resource view
				m.State = stateResourceView
				return m, nil
			case stateLoadBalancerTargetView:
				// From load balancer target view, go back to resource view 
				m.State = stateResourceView 
				return m, nil 
			case stateLoadBalancerServiceView:
				// From load balancer service view, go back to resource view 
				m.State = stateResourceView 
				return m, nil 
			
			}
		}

		// Handle state-specific keys (excluding quit, which is handled above)
		switch m.State {
		case StateProjectSelect:
			switch {
			case key.Matches(msg, keys.Enter):
				if selectedItem := m.projectList.SelectedItem(); selectedItem != nil {
					if projectItem, ok := selectedItem.(r_prj.ProjectItem); ok {
						m.client = hcloud.NewClient(hcloud.WithToken(projectItem.Config.Token))
						m.currentProject = projectItem.Config.Name
						m.State = stateResourceView
						// Reset loaded resources for new project
						m.LoadedResources = make(map[resource.ResourceType]bool)
						// Load the first tab's resources
						return m, tea.Batch(
							resource.StartResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			case key.Matches(msg, keys.Add):
				m.projectForm = project.NewProjectForm()
				m.State = stateProjectManage

			case key.Matches(msg, keys.Delete):
				if selectedItem := m.projectList.SelectedItem(); selectedItem != nil {
					if projectItem, ok := selectedItem.(r_prj.ProjectItem); ok {
						m.config.RemoveProject(projectItem.Config.Name)
						m.updateProjectList()
						return m, config.SaveConfigCmd(m.config)
					}
				}
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
					m.State = StateProjectSelect
					return m, config.SaveConfigCmd(m.config)
				}
			}

		case stateTokenInput:
			switch {
			case key.Matches(msg, keys.Enter):
				token := strings.TrimSpace(m.TokenInput.Value())
				if token == "" {
					return m, nil
				}

				m.client = hcloud.NewClient(hcloud.WithToken(token))
				m.State = stateResourceView
				// Reset loaded resources
				m.LoadedResources = make(map[resource.ResourceType]bool)
				// Load the first tab's resources
				return m, tea.Batch(
					resource.StartResourceLoad(m.activeTab),
					m.getResourceLoadCmd(m.activeTab),
				)
			}

		case stateResourceView:
			switch {
			case key.Matches(msg, keys.Enter):
				// Show context menu for servers
				switch m.activeTab {
				case resource.ResourceServers:
					if currentList, exists := m.Lists[resource.ResourceServers]; exists {
						if selectedItem := currentList.SelectedItem(); selectedItem != nil {
							if serverItem, ok := selectedItem.(r_serv.ServerItem); ok {
								m.contextMenu = ctm_serv.CreateServerContextMenu(serverItem.Server) 
								m.State = stateContextMenu
							}
						}
					}
				case resource.ResourceNetworks:
					if currentList, exists := m.Lists[resource.ResourceNetworks]; exists {
						if selectedItem := currentList.SelectedItem(); selectedItem != nil {
							if networkItem, ok := selectedItem.(r_n.NetworkItem); ok {
								m.contextMenu = ctm_n.CreateNetworkContextMenu(networkItem.Network)
								m.State = stateContextMenu
							}
						}
					}

				case resource.ResourceLoadBalancers:
					if currentList, exists := m.Lists[resource.ResourceLoadBalancers]; exists {
						if selectedItem := currentList.SelectedItem(); selectedItem != nil {
							if lbItem, ok := selectedItem.(r_lb.LoadBalancerItem); ok {
								m.contextMenu = ctm_lb.CreateLoadbalancerContextMenu(lbItem.Lb)
								m.State = stateContextMenu
							}
						}
					}
}				

			case key.Matches(msg, keys.Tab):
				m.activeTab = (m.activeTab + 1) % resource.ResourceType(len(resourceTabs))

				// Load resources for new tab if not already loaded
				if !m.LoadedResources[m.activeTab] {
					return m, tea.Batch(
						resource.StartResourceLoad(m.activeTab),
						m.getResourceLoadCmd(m.activeTab),
					)
				}
			case key.Matches(msg, keys.Reload):
				if m.client != nil {
					// Mark current resource as not loaded and reload
					m.LoadedResources[m.activeTab] = false
					return m, tea.Batch(
						resource.StartResourceLoad(m.activeTab),
						m.getResourceLoadCmd(m.activeTab),
					)
				}
			case key.Matches(msg, keys.Left):
				if m.activeTab > 0 {
					m.activeTab--

					// Load resources for new tab if not already loaded
					if !m.LoadedResources[m.activeTab] {
						return m, tea.Batch(
							resource.StartResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			case key.Matches(msg, keys.Right):
				if int(m.activeTab) < len(resourceTabs)-1 {
					m.activeTab++

					// Load resources for new tab if not already loaded
					if !m.LoadedResources[m.activeTab] {
						return m, tea.Batch(
							resource.StartResourceLoad(m.activeTab),
							m.getResourceLoadCmd(m.activeTab),
						)
					}
				}
			}

		case stateLabelView:
			// Label view specific keys (quit is handled globally above)
			// Add any other label view specific key handling here
			break

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
				m.State = stateResourceView
				return m, m.executeContextAction(selectedAction, m.contextMenu.ResourceType, m.contextMenu.ResourceID) 

			case key.Matches(msg, keys.Num1, keys.Num2, keys.Num3, keys.Num4, keys.Num5,
				keys.Num6, keys.Num7, keys.Num8, keys.Num9, keys.Num0):

				keyStr := msg.String()
				selectedIndex := util.GetIndexFromNumber(keyStr)

				// Check if the number corresponds to a valid menu item
				if selectedIndex >= 0 && selectedIndex < len(m.contextMenu.Items) {
					selectedAction := m.contextMenu.Items[selectedIndex].Action
					m.State = stateResourceView
					return m, m.executeContextAction(selectedAction, m.contextMenu.ResourceType, m.contextMenu.ResourceID)
				}
			}

		case stateError:
			// Error state - quit is handled globally above
			break
		}

	case config.ProjectSavedMsg:
		m.statusMessage = "✅ Project configuration saved"
		return m, clearStatusMessage()

	case resource.ResourceLoadStartMsg:
		m.IsLoading = true
		m.loadingResource = msg.ResourceType
		return m, nil

	case r_serv.ServersLoadedMsg:
		m.IsLoading = false
		m.LoadedResources[resource.ResourceServers] = true

		// Create server list
		serverItems := make([]list.Item, len(msg.Servers))
		for i, server := range msg.Servers {
			serverItems[i] = r_serv.ServerItem{
				Server: server,
				ResourceType: resource.ResourceServers,
				ResourceID:   server.ID,
			}
		}

		serversList := list.New(serverItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		serversList.Title = "Servers"
		m.Lists[resource.ResourceServers] = serversList
		return m, nil

	case r_n.NetworksLoadedMsg:
		m.IsLoading = false
		m.LoadedResources[resource.ResourceNetworks] = true

		// Create network list
		networkItems := make([]list.Item, len(msg.Networks))
		for i, network := range msg.Networks {
			networkItems[i] = r_n.NetworkItem{
				Network: network,
				ResourceType: resource.ResourceNetworks,
				ResourceID:   network.ID,
			}
		}

		networksList := list.New(networkItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		networksList.Title = "Networks"
		m.Lists[resource.ResourceNetworks] = networksList
		return m, nil

	case r_lb.LoadBalancersLoadedMsg:
		m.IsLoading = false
		m.LoadedResources[resource.ResourceLoadBalancers] = true

		// Create load balancer list
		lbItems := make([]list.Item, len(msg.LoadBalancers))
		for i, lb := range msg.LoadBalancers {
			lbItems[i] = r_lb.LoadBalancerItem{
				Lb: lb,
				ResourceType: resource.ResourceLoadBalancers,
				ResourceID:   lb.ID,
			}
		}

		lbList := list.New(lbItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		lbList.Title = "Load Balancers"
		m.Lists[resource.ResourceLoadBalancers] = lbList
	        return m, nil
	
	case r_vol.VolumesLoadedMsg:
		m.IsLoading = false
		m.LoadedResources[resource.ResourceVolumes] = true

		// Create volume list
		volumeItems := make([]list.Item, len(msg.Volumes))
		for i, volume := range msg.Volumes {
			volumeItems[i] = r_vol.VolumeItem{Volume: volume}
		}

		volumesList := list.New(volumeItems, list.NewDefaultDelegate(), m.width-4, m.height-10)
		volumesList.Title = "Volumes"
		m.Lists[resource.ResourceVolumes] = volumesList
		return m, nil

	case message.ClipboardCopiedMsg:
		m.statusMessage = fmt.Sprintf("✅ Copied %s to clipboard", string(msg))
		return m, clearStatusMessage()

	case ctm_serv.SshLaunchedMsg:
		m.statusMessage = "🚀 SSH session launched"
		return m, clearStatusMessage()
	case ctm_serv.TmuxSSHLaunchedMsg:
		m.statusMessage = "🪟 SSH session launched in tmux"
		return m, clearStatusMessage()

	case ctm_serv.ZellijSSHLaunchedMsg:
		m.statusMessage = "📱 SSH session launched in zellij"
		return m, clearStatusMessage()

	case r_label.LabelsLoadedMsg:
		m.IsLoading = false
		m.loadedLabels = msg.Labels
		m.labelsPertainingToResource = msg.RelatedResourceName
		m.State = stateLabelView
		return m, nil

	case r_lb.ViewLoadbalancerTargetsMsg:
		m.IsLoading = false
		m.loadbalancerBeingViewed = msg.LoadBalancer
		m.loadbalancerTargets = msg.Targets
		m.State = stateLoadBalancerTargetView 
		return m, nil

	case r_lb.ViewLoadbalancerServicesMsg:
		m.IsLoading = false 
		m.loadbalancerBeingViewed = msg.LoadBalancer 
		m.loadbalancerServices = msg.Services 
		m.State = stateLoadBalancerServiceView 
		return m, nil 

	case message.CancelCtxMenuMsg:
		// close the context menu and return to resource view
		m.State = stateResourceView
		return m, nil

	case message.StatusMsg:
		m.statusMessage = string(msg)
		return m, clearStatusMessage()

	case message.ErrorMsg:
		m.State = stateError
		m.err = msg.Err
		return m, nil
	}

	// Update components
	if m.State == stateTokenInput {
		var cmd tea.Cmd
		m.TokenInput, cmd = m.TokenInput.Update(msg)
		return m, cmd
	}

	if m.State == StateProjectSelect && m.config != nil {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		return m, cmd
	}

	if m.State == stateProjectManage {
		var cmd tea.Cmd
		m.projectForm.Inputs[m.projectForm.FocusIdx], cmd = m.projectForm.Inputs[m.projectForm.FocusIdx].Update(msg)
		return m, cmd
	}

	if m.State == stateResourceView {
		if currentList, exists := m.Lists[m.activeTab]; exists {
			var cmd tea.Cmd
			m.Lists[m.activeTab], cmd = currentList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}
