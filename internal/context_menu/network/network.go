package network

import (
	"fmt"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	ctm "github.com/grammeaway/lazyhetzner/internal/context_menu"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/grammeaway/lazyhetzner/internal/resource/label"
	r_n "github.com/grammeaway/lazyhetzner/internal/resource/network"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func CreateNetworkContextMenu(network *hcloud.Network) ctm.ContextMenu {
	return ctm.ContextMenu{
		Items:        getNetworkMenuItems(),
		SelectedItem: 0,
		ResourceType: resource.ResourceNetworks,
		ResourceID:   network.ID,
	}
}

func getNetworkLabels(network *hcloud.Network) map[string]string {
	labels := make(map[string]string)
	for key, value := range network.Labels {
		labels[key] = value
	}
	return labels
}

// Returns context menu Items for networks
func getNetworkMenuItems() []ctm.ContextMenuItem {
	return []ctm.ContextMenuItem{
		// add action for canceling (i.e., closing) the context menu
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🧩 View Subnets", Action: "view_subnets"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		// Copy the network ID to clipboard
		{Label: "📋 Copy Network ID", Action: "copy_id"},
		// Copy the network name to clipboard
		{Label: "📋 Copy Network Name", Action: "copy_name"},
		// copy CIDR to clipboard
		{Label: "📋 Copy IP Range", Action: "copy_ip_range"},
	}
}

func ExecuteNetworkContextAction(selectedAction string, network *hcloud.Network) tea.Cmd {
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_subnets":
		if len(network.Subnets) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No subnets found for this network.")
			}
		}
		return func() tea.Msg {
			return r_n.ViewNetworkSubnetsMsg{
				Network: network,
				Subnets: network.Subnets,
			}
		}
	case "view_labels":
		labels := getNetworkLabels(network)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this server.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Network: %s", network.Name),
				RelatedResourceType: resource.ResourceServers,
			}
		}
	case "copy_id":
		// Copy the network ID to clipboard
		return func() tea.Msg {
			if err := clipboard.WriteAll(fmt.Sprintf("%d", network.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Network ID %d copied to clipboard", network.ID))
		}
	case "copy_name":
		// Copy the network name to clipboard
		return func() tea.Msg {
			if network.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("network has no name")}
			}
			if err := clipboard.WriteAll(network.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Network name '%s' copied to clipboard", network.Name))
		}
	case "copy_ip_range":
		// Copy the IP range to clipboard
		return func() tea.Msg {
			if network.IPRange.String() == "" {
				return message.ErrorMsg{Err: fmt.Errorf("network has no IP range")}
			}
			if err := clipboard.WriteAll(network.IPRange.String()); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("IP range '%s' copied to clipboard", network.IPRange))
		}

	default:
		return nil
	}
}
