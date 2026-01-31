package model

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	util "github.com/grammeaway/lazyhetzner/utility"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"net"
	"sort"
	"strings"
)

func (m Model) View() string {
	switch m.State {
	case StateProjectSelect:
		if m.config == nil {
			return fmt.Sprintf(
				"\n%s\n\n%s\n",
				titleStyle.Render("lazyhetzner"),
				infoStyle.Render("Loading configuration..."),
			)
		}

		if len(m.config.Projects) == 0 {
			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s\n",
				titleStyle.Render("lazyhetzner"),
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
			titleStyle.Render("lazyhetzner"),
			m.projectList.View(),
			statusView,
			helpStyle.Render("Enter: select project • a: add project • d: delete project • q: quit"),
		)

	case stateProjectManage:
		var formView strings.Builder
		formView.WriteString("Add New Project\n\n")

		for i, input := range m.projectForm.Inputs {
			var style lipgloss.Style
			if i == m.projectForm.FocusIdx {
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
			titleStyle.Render("lazyhetzner - Add Project"),
			formView.String(),
		)
	case stateTokenInput:
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n\n%s\n",
			titleStyle.Render("lazyhetzner"),
			infoStyle.Render("Enter API token for one-time access:"),
			m.TokenInput.View(),
			helpStyle.Render("Press Enter to continue • Press q to go back"),
		)

	case stateLoading:
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			titleStyle.Render("lazyhetzner"),
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
			if resource.ResourceType(i) == m.activeTab {
				// Show loading indicator in active tab if loading
				if m.IsLoading && m.loadingResource == resource.ResourceType(i) {
					tabs = append(tabs, titleStyle.Render(tab+" (loading...)"))
				} else {
					tabs = append(tabs, titleStyle.Render(tab))
				}
			} else {
				tabs = append(tabs, helpStyle.Render(tab))
			}
		}
		tabsView := strings.Join(tabs, " ")

		// Render current list or loading message
		var listView string
		if m.IsLoading && m.loadingResource == m.activeTab {
			listView = infoStyle.Render("Loading " + strings.ToLower(resourceTabs[m.activeTab]) + "...")
		} else if currentList, exists := m.Lists[m.activeTab]; exists {
			listView = currentList.View()
		} else if !m.LoadedResources[m.activeTab] {
			listView = helpStyle.Render("Resources not loaded yet. Loading will start automatically.")
		} else {
			listView = helpStyle.Render("No " + strings.ToLower(resourceTabs[m.activeTab]) + " found.")
		}

		// Status message
		statusView := ""
		if m.statusMessage != "" {
			statusView = "\n" + successStyle.Render(m.statusMessage)
		}

		helpText := "Tab: switch view • ←/→: navigate tabs • Enter: actions • r: reload resources • q: back to projects"
		if m.activeTab == resource.ResourceServers {
			helpText = "Tab: switch view • ←/→: navigate tabs • Enter: server actions • i: view details • r: reload resources • q: back to projects"
		}

		return fmt.Sprintf(
			"%s\n%s\n\n%s%s\n\n%s",
			infoStyle.Render(projectHeader),
			tabsView,
			listView,
			statusView,
			helpStyle.Render(helpText),
		)
	case stateLoadBalancerServiceView:
		// Render the Load Balancer services, as a list of loadbalancerServices
		var serviceView strings.Builder
		serviceView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Load Balancer Services")))
		resourceInfo := fmt.Sprintf("🔌Services for Load Balancer: %s", m.loadbalancerBeingViewed.Name)
		serviceView.WriteString(infoStyle.Render(resourceInfo) + "\n\n")
		if len(m.loadbalancerServices) == 0 {
			noServicesMsg := "⚠️  No services found for this Load Balancer"
			serviceView.WriteString(noServicesStyle.Render(noServicesMsg) + "\n")
		} else {
			// Create a container for all loadbalancerServices
			var servicesContent strings.Builder
			servicesContent.WriteString(fmt.Sprintf("Found %d service(s):\n\n", len(m.loadbalancerServices)))
			// Render each service with improved styling
			for i, service := range m.loadbalancerServices {
				// Create a styled service description
				serviceDesc := fmt.Sprintf("🔌Service %d: %s (Protocol: %s, Port: %d -> %d)", i+1, service.Protocol, service.Protocol, service.ListenPort, service.DestinationPort)
				serviceStyled := serviceStyle.Render(serviceDesc)
				// Add the service to the content
				servicesContent.WriteString(serviceStyled + "\n")
				// Add spacing between services (except for the last one)
				if i < len(m.loadbalancerServices)-1 {
					servicesContent.WriteString("\n")
				}
			}
			// Wrap all services in a container
			serviceView.WriteString(serviceContainerStyle.Render(servicesContent.String()) + "\n")
		}
		// Enhanced help text
		helpText := "💡 Press 'q' to return to Load Balancer view"
		serviceView.WriteString("\n" + helpStyle.Render(helpText))
		return serviceView.String()

	case stateLoadBalancerTargetView:
		// Render the Load Balancer targets, as a list of loadbalancerTargets
		var targetView strings.Builder
		targetView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Load Balancer Targets")))
		resourceInfo := fmt.Sprintf("🎯Targets for Load Balancer: %s", m.loadbalancerBeingViewed.Name)
		targetView.WriteString(infoStyle.Render(resourceInfo) + "\n\n")
		if len(m.loadbalancerTargets) == 0 {
			noTargetsMsg := "⚠️  No targets found for this Load Balancer"
			targetView.WriteString(noTargetsStyle.Render(noTargetsMsg) + "\n")
		} else {
			// Create a container for all loadbalancerTargets
			var targetsContent strings.Builder
			targetsContent.WriteString(fmt.Sprintf("Found %d target(s):\n\n", len(m.loadbalancerTargets)))
			// Render each target with improved styling
			for i, target := range m.loadbalancerTargets {
				// Create a styled target description
				targetDesc := fmt.Sprintf("🎯Target %d: %s (Type: %s, Target count: %d)", i+1, target.LabelSelector.Selector, target.Type, len(target.Targets))
				targetStyled := targetStyle.Render(targetDesc)
				// Add the target to the content
				targetsContent.WriteString(targetStyled + "\n")
				// Add spacing between targets (except for the last one)
				if i < len(m.loadbalancerTargets)-1 {
					targetsContent.WriteString("\n")
				}
			}
			// Wrap all targets in a container
			targetView.WriteString(targetContainerStyle.Render(targetsContent.String()) + "\n")
		}
		// Enhanced help text
		helpText := "💡 Press 'q' to return to Load Balancer view"
		targetView.WriteString("\n" + helpStyle.Render(helpText))
		return targetView.String()
	case stateFirewallRuleView:
		var ruleView strings.Builder
		ruleView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Firewall Rules")))
		resourceInfo := fmt.Sprintf("🧱 Rules for Firewall: %s", m.firewallBeingViewed.Name)
		ruleView.WriteString(infoStyle.Render(resourceInfo) + "\n\n")
		if len(m.firewallRules) == 0 {
			noRulesMsg := "⚠️  No rules found for this Firewall"
			ruleView.WriteString(noFirewallRulesStyle.Render(noRulesMsg) + "\n")
		} else {
			var rulesContent strings.Builder
			rulesContent.WriteString(fmt.Sprintf("Found %d rule(s):\n\n", len(m.firewallRules)))
			for i, rule := range m.firewallRules {
				ruleDesc := fmt.Sprintf("🧱 Rule %d: %s %s %s", i+1, strings.ToUpper(string(rule.Direction)), strings.ToUpper(string(rule.Protocol)), formatFirewallPort(rule.Port))
				ruleDetails := fmt.Sprintf("Sources: %s | Destinations: %s", formatIPNets(rule.SourceIPs), formatIPNets(rule.DestinationIPs))
				if rule.Description != nil && *rule.Description != "" {
					ruleDetails = fmt.Sprintf("%s | %s", ruleDetails, *rule.Description)
				}
				ruleStyled := firewallRuleStyle.Render(ruleDesc + "\n" + ruleDetails)
				rulesContent.WriteString(ruleStyled + "\n")
				if i < len(m.firewallRules)-1 {
					rulesContent.WriteString("\n")
				}
			}
			ruleView.WriteString(firewallRuleContainerStyle.Render(rulesContent.String()) + "\n")
		}
		helpText := "💡 Press 'q' to return to Firewall view"
		ruleView.WriteString("\n" + helpStyle.Render(helpText))
		return ruleView.String()

	case stateNetworkSubnetView:
		var subnetView strings.Builder
		subnetView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Network Subnets")))
		resourceInfo := fmt.Sprintf("🧩 Subnets for Network: %s", m.networkBeingViewed.Name)
		subnetView.WriteString(infoStyle.Render(resourceInfo) + "\n\n")
		if len(m.networkSubnets) == 0 {
			noSubnetsMsg := "⚠️  No subnets found for this Network"
			subnetView.WriteString(noSubnetsStyle.Render(noSubnetsMsg) + "\n")
		} else {
			var subnetsContent strings.Builder
			subnetsContent.WriteString(fmt.Sprintf("Found %d subnet(s):\n\n", len(m.networkSubnets)))
			for i, subnet := range m.networkSubnets {
				subnetDesc := fmt.Sprintf("🧩 Subnet %d: %s (%s)", i+1, subnet.IPRange.String(), strings.ToUpper(string(subnet.Type)))
				subnetDetails := fmt.Sprintf("Network Zone: %s | Gateway: %s", subnet.NetworkZone, formatSubnetGateway(subnet.Gateway))
				subnetStyled := subnetStyle.Render(subnetDesc + "\n" + subnetDetails)
				subnetsContent.WriteString(subnetStyled + "\n")
				if i < len(m.networkSubnets)-1 {
					subnetsContent.WriteString("\n")
				}
			}
			subnetView.WriteString(subnetContainerStyle.Render(subnetsContent.String()) + "\n")
		}
		helpText := "💡 Press 'q' to return to Network view"
		subnetView.WriteString("\n" + helpStyle.Render(helpText))
		return subnetView.String()

	case stateServerDetailView:
		if m.serverBeingViewed == nil {
			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s\n",
				titleStyle.Render("Server Details"),
				warningStyle.Render("No server details available."),
				helpStyle.Render("Press 'q' to return to resource view"),
			)
		}

		server := m.serverBeingViewed
		var detailView strings.Builder
		detailView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Server Details")))
		header := fmt.Sprintf("🖥️ Server: %s (ID: %d)", server.Name, server.ID)
		detailView.WriteString(infoStyle.Render(header) + "\n\n")

		overviewLines := []string{
			fmt.Sprintf("Status: %s", server.Status),
			fmt.Sprintf("Type: %s", formatServerType(server)),
			fmt.Sprintf("Datacenter: %s", formatDatacenter(server)),
			fmt.Sprintf("Image: %s", formatServerImage(server)),
			fmt.Sprintf("Created: %s", server.Created.Format("2006-01-02 15:04:05")),
			fmt.Sprintf("Rescue Enabled: %t", server.RescueEnabled),
		}
		if server.PlacementGroup != nil {
			overviewLines = append(overviewLines, fmt.Sprintf("Placement Group: %s", server.PlacementGroup.Name))
		}
		overviewSection := renderServerDetailSection("Overview", overviewLines)

		networkLines := []string{
			fmt.Sprintf("Public IPv4: %s", formatIP(server.PublicNet.IPv4.IP)),
			fmt.Sprintf("Public IPv6: %s", formatIP(server.PublicNet.IPv6.IP)),
			fmt.Sprintf("Floating IPs: %s", formatFloatingIPs(server.PublicNet.FloatingIPs)),
		}
		privateNetLines := formatPrivateNetworks(server)
		if len(privateNetLines) > 0 {
			networkLines = append(networkLines, "Private Networks:")
			networkLines = append(networkLines, privateNetLines...)
		}
		networkSection := renderServerDetailSection("Networking", networkLines)
		subnetSection := renderServerDetailSection("Subnets", formatServerSubnets(m.serverDetailNetworks))
		firewallSection := renderServerDetailSection("Firewalls", formatServerFirewalls(server))
		loadBalancerSection := renderServerDetailSection("Load Balancers", formatServerLoadBalancers(server))
		volumeSection := renderServerDetailSection("Volumes", formatServerVolumes(server))
		labelSection := renderServerDetailSection("Labels", formatServerLabels(server))

		sections := []string{
			overviewSection,
			networkSection,
			subnetSection,
			firewallSection,
			loadBalancerSection,
			volumeSection,
			labelSection,
		}

		detailView.WriteString(renderServerDetailGrid(sections, m.width) + "\n")

		helpText := "💡 Press 'q' to return to resource view"
		detailView.WriteString(helpStyle.Render(helpText))
		return detailView.String()

	case stateLabelView:
		// Render the label View
		var labelView strings.Builder
		labelView.WriteString(fmt.Sprintf("%s\n\n", titleStyle.Render("Labels")))

		// Resource info with better styling
		resourceInfo := fmt.Sprintf("📋 Labels for %s", m.labelsPertainingToResource)
		labelView.WriteString(infoStyle.Render(resourceInfo) + "\n\n")

		if len(m.loadedLabels) == 0 {
			// Enhanced no labels message
			noLabelsMsg := "⚠️  No labels found for this resource"
			labelView.WriteString(noLabelsStyle.Render(noLabelsMsg) + "\n")
		} else {
			// Create a container for all labels
			var labelsContent strings.Builder
			labelsContent.WriteString(fmt.Sprintf("Found %d label(s):\n\n", len(m.loadedLabels)))

			// Sort keys for consistent display (optional)
			keys := make([]string, 0, len(m.loadedLabels))
			for k := range m.loadedLabels {
				keys = append(keys, k)
			}

			// Render each label with improved styling
			for i, key := range keys {
				value := m.loadedLabels[key]

				// Create a styled key-value pair
				keyStyled := labelKeyStyle.Render("🏷️  " + key)
				valueStyled := labelValueStyle.Render(value)

				// Join key and value with some spacing
				labelPair := lipgloss.JoinHorizontal(
					lipgloss.Center,
					keyStyled,
					" → ",
					valueStyled,
				)

				labelsContent.WriteString(labelPair)

				// Add spacing between labels (except for the last one)
				if i < len(keys)-1 {
					labelsContent.WriteString("\n\n")
				}
			}

			// Wrap all labels in a container
			labelView.WriteString(labelContainerStyle.Render(labelsContent.String()) + "\n")
		}

		// Enhanced help text
		helpText := "💡 Press 'q' to return to resource view"
		labelView.WriteString("\n" + helpStyle.Render(helpText))

		return labelView.String()
	case stateContextMenu:
		// Render the current resource view in background
		projectHeader := fmt.Sprintf("Project: %s", m.currentProject)
		if m.currentProject == "" {
			projectHeader = "One-time Access"
		}

		var tabs []string
		for i, tab := range resourceTabs {
			if resource.ResourceType(i) == m.activeTab {
				tabs = append(tabs, titleStyle.Render(tab))
			} else {
				tabs = append(tabs, helpStyle.Render(tab))
			}
		}
		tabsView := strings.Join(tabs, " ")

		var listView string
		if currentList, exists := m.Lists[m.activeTab]; exists {
			listView = currentList.View()
		}

		// Render context menu with number shortcuts
		var menuItems []string
		for i, item := range m.contextMenu.Items {
			// Get the number for this item (1-indexed, with 0 for 10th)
			numberStr := util.GetNumberForIndex(i)

			// Create the menu item with number prefix
			menuText := fmt.Sprintf("[%s] %s", numberStr, item.Label)

			if i == m.contextMenu.SelectedItem {
				menuItems = append(menuItems, selectedMenuStyle.Render(menuText))
			} else {
				menuItems = append(menuItems, menuText)
			}
		}

		menuContent := strings.Join(menuItems, "\n")

		// Add help text about number shortcuts
		helpText := "\nPress number keys for quick selection • ↑/↓ to navigate • Enter to select • Esc to cancel"
		menu := menuStyle.Render(fmt.Sprintf("Actions for %s:\n\n%s%s",

			resource.GetResourceNameFromType(m.contextMenu.ResourceType),

			menuContent,
			helpStyle.Render(helpText)))

		// Center the menu
		menuOverlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menu)

		background := fmt.Sprintf("%s\n%s\n\n%s", infoStyle.Render(projectHeader), tabsView, listView)

		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, background) + menuOverlay
	case stateError:
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			titleStyle.Render("lazyhetzner - Error"),
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
			helpStyle.Render("Press q to quit"),
		)
	}

	return ""
}

func formatFirewallPort(port *string) string {
	if port == nil || *port == "" {
		return "all ports"
	}
	return fmt.Sprintf("port %s", *port)
}

func formatIPNets(nets []net.IPNet) string {
	if len(nets) == 0 {
		return "any"
	}
	parts := make([]string, 0, len(nets))
	for _, ipNet := range nets {
		parts = append(parts, ipNet.String())
	}
	return strings.Join(parts, ", ")
}

func formatSubnetGateway(gateway net.IP) string {
	if gateway == nil || len(gateway) == 0 {
		return "n/a"
	}
	return gateway.String()
}

func renderServerDetailSection(title string, lines []string) string {
	if len(lines) == 0 {
		lines = []string{"No data available."}
	}
	content := strings.Join(lines, "\n")
	return serverDetailSectionStyle.Render(serverDetailTitleStyle.Render(title) + "\n" + content)
}

func renderServerDetailGrid(sections []string, width int) string {
	if len(sections) == 0 {
		return ""
	}
	gridWidth := max(40, width-4)
	columns := 2
	if gridWidth < 90 {
		columns = 1
	}
	columnWidth := gridWidth
	if columns == 2 {
		columnWidth = max(30, (gridWidth-2)/2)
	}
	cellStyle := lipgloss.NewStyle().Width(columnWidth)

	rows := make([]string, 0, (len(sections)+columns-1)/columns)
	for i := 0; i < len(sections); i += columns {
		rowCells := make([]string, 0, columns)
		for j := 0; j < columns; j++ {
			idx := i + j
			if idx < len(sections) {
				rowCells = append(rowCells, cellStyle.Render(sections[idx]))
			} else if columns == 2 {
				rowCells = append(rowCells, cellStyle.Render(""))
			}
		}
		if columns == 1 {
			rows = append(rows, rowCells[0])
		} else {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
		}
	}

	return strings.Join(rows, "\n")
}

func formatIP(ip net.IP) string {
	if ip == nil || len(ip) == 0 {
		return "n/a"
	}
	return ip.String()
}

func formatServerType(server *hcloud.Server) string {
	if server.ServerType == nil {
		return "n/a"
	}
	return server.ServerType.Name
}

func formatDatacenter(server *hcloud.Server) string {
	if server.Datacenter == nil {
		return "n/a"
	}
	return server.Datacenter.Name
}

func formatServerImage(server *hcloud.Server) string {
	if server.Image == nil {
		return "n/a"
	}
	if server.Image.Name != "" {
		return server.Image.Name
	}
	return fmt.Sprintf("Image %d", server.Image.ID)
}

func formatFloatingIPs(floatingIPs []*hcloud.FloatingIP) string {
	if len(floatingIPs) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(floatingIPs))
	for _, floatingIP := range floatingIPs {
		if floatingIP == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (ID: %d)", floatingIP.IP, floatingIP.ID))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func formatPrivateNetworks(server *hcloud.Server) []string {
	if server == nil || len(server.PrivateNet) == 0 {
		return []string{"No private networks attached."}
	}
	lines := make([]string, 0, len(server.PrivateNet))
	for _, privateNet := range server.PrivateNet {
		networkName := "unknown"
		networkID := int64(0)
		if privateNet.Network != nil {
			networkName = privateNet.Network.Name
			networkID = privateNet.Network.ID
		}
		aliasText := "none"
		if len(privateNet.Aliases) > 0 {
			aliases := make([]string, 0, len(privateNet.Aliases))
			for _, alias := range privateNet.Aliases {
				if alias != nil {
					aliases = append(aliases, alias.String())
				}
			}
			if len(aliases) > 0 {
				aliasText = strings.Join(aliases, ", ")
			}
		}
		lines = append(lines, fmt.Sprintf("• %s (ID: %d) | IP: %s | MAC: %s | Aliases: %s", networkName, networkID, formatIP(privateNet.IP), privateNet.MACAddress, aliasText))
	}
	return lines
}

func formatServerSubnets(networks []*hcloud.Network) []string {
	if len(networks) == 0 {
		return []string{"No subnets available."}
	}
	lines := []string{}
	for _, network := range networks {
		if network == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s (ID: %d)", network.Name, network.ID))
		if len(network.Subnets) == 0 {
			lines = append(lines, "  • No subnets defined.")
			continue
		}
		for _, subnet := range network.Subnets {
			ipRange := "n/a"
			if subnet.IPRange != nil {
				ipRange = subnet.IPRange.String()
			}
			lines = append(lines, fmt.Sprintf("  • %s (%s) | Zone: %s | Gateway: %s", ipRange, strings.ToUpper(string(subnet.Type)), subnet.NetworkZone, formatSubnetGateway(subnet.Gateway)))
		}
	}
	return lines
}

func formatServerFirewalls(server *hcloud.Server) []string {
	if server == nil || len(server.PublicNet.Firewalls) == 0 {
		return []string{"No firewalls attached."}
	}
	lines := make([]string, 0, len(server.PublicNet.Firewalls))
	for _, firewallStatus := range server.PublicNet.Firewalls {
		lines = append(lines, fmt.Sprintf("• %s (ID: %d) | Status: %s", firewallStatus.Firewall.Name, firewallStatus.Firewall.ID, firewallStatus.Status))
	}
	return lines
}

func formatServerLoadBalancers(server *hcloud.Server) []string {
	if server == nil || len(server.LoadBalancers) == 0 {
		return []string{"No load balancers attached."}
	}
	lines := make([]string, 0, len(server.LoadBalancers))
	for _, lb := range server.LoadBalancers {
		if lb == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("• %s (ID: %d)", lb.Name, lb.ID))
	}
	return lines
}

func formatServerVolumes(server *hcloud.Server) []string {
	if server == nil || len(server.Volumes) == 0 {
		return []string{"No volumes attached."}
	}
	lines := make([]string, 0, len(server.Volumes))
	for _, volume := range server.Volumes {
		if volume == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("• %s (ID: %d) | Size: %d GB", volume.Name, volume.ID, volume.Size))
	}
	return lines
}

func formatServerLabels(server *hcloud.Server) []string {
	if server == nil || len(server.Labels) == 0 {
		return []string{"No labels attached."}
	}
	keys := make([]string, 0, len(server.Labels))
	for key := range server.Labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("• %s=%s", key, server.Labels[key]))
	}
	return lines
}
