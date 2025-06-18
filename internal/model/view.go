package model

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"lazyhetzner/internal/resource"
	util "lazyhetzner/utility"
	"lazyhetzner/internal/input_form/server/snapshot"
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
			helpText = "Tab: switch view • ←/→: navigate tabs • Enter: server actions • r: reload resources • q: back to projects"
		}

		return fmt.Sprintf(
			"%s\n%s\n\n%s%s\n\n%s",
			infoStyle.Render(projectHeader),
			tabsView,
			listView,
			statusView,
			helpStyle.Render(helpText),
		)

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
