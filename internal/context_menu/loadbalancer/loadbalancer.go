package loadbalancer 

import (
	"fmt"
	ctm "lazyhetzner/internal/context_menu"
	"lazyhetzner/internal/message"
	"lazyhetzner/internal/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"lazyhetzner/internal/resource/label"
)

func CreateLoadbalancerContextMenu(loadbalancer *hcloud.LoadBalancer) ctm.ContextMenu {
	return ctm.ContextMenu{
		Items:        getLoadbalancerMenuItems(),
		SelectedItem: 0,
		ResourceType: resource.ResourceLoadBalancers,
		ResourceID:   loadbalancer.ID,
	}
}




func getLoadbalancerLabels(loadbalancer *hcloud.LoadBalancer) map[string]string {
	labels := make(map[string]string)
	for key, value := range loadbalancer.Labels {
		labels[key] = value
	}
	return labels
}



// Returns context menu Items for loadbalancers
func getLoadbalancerMenuItems() []ctm.ContextMenuItem {
	return []ctm.ContextMenuItem{
		// add action for canceling (i.e., closing) the context menu
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		// Copy the loadbalancer ID to clipboard
		{Label: "📋 Copy Loadbalancer ID", Action: "copy_id"},
		// Copy the loadbalancer name to clipboard
		{Label: "📋 Copy Loadbalancer Name", Action: "copy_name"},
		{Label: "📋 Copy Public IP (IPv4)", Action: "copy_public_ip"},
		{ Label: "📋 Copy Public IP (IPv6)", Action: "copy_public_ipv6"}, // Assuming IPv6 is also available
  		// Copy private IP to clipboard
		{Label: "📋 Copy Private IP", Action: "copy_private_ip"},


	}
}



func ExecuteLoadbalancerContextAction(selectedAction string, loadbalancer *hcloud.LoadBalancer) tea.Cmd {
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_labels":
		labels := getLoadbalancerLabels(loadbalancer)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this server.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Loadbalancer: %s", loadbalancer.Name),
				RelatedResourceType: resource.ResourceServers,
			}
		}
	case "copy_id":
		// Copy the loadbalancer ID to clipboard
		return func() tea.Msg {
			if err := clipboard.WriteAll(fmt.Sprintf("%d", loadbalancer.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("%d", loadbalancer.ID))

		}
	case "copy_name":
		// Copy the loadbalancer name to clipboard
		return func() tea.Msg {
			if loadbalancer.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("loadbalancer has no name")}
			}
			if err := clipboard.WriteAll(loadbalancer.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(loadbalancer.Name)
		}
	case "copy_public_ip":
		// Copy the public IP to clipboard
		return func() tea.Msg {
			if loadbalancer.PublicNet.IPv4.IP == nil {
				return message.ErrorMsg{Err: fmt.Errorf("loadbalancer has no public IP")}
			}
			if err := clipboard.WriteAll(loadbalancer.PublicNet.IPv4.IP.String()); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(loadbalancer.PublicNet.IPv4.IP.String())
		}
	case "copy_public_ipv6":
		// Copy the public IPv6 to clipboard
		if loadbalancer.PublicNet.IPv6.IP == nil {
			return func() tea.Msg {
				return message.ErrorMsg{Err: fmt.Errorf("loadbalancer has no public IPv6")}
			}
		}
		if err := clipboard.WriteAll(loadbalancer.PublicNet.IPv6.IP.String()); err != nil {
			return func() tea.Msg {
				return message.ErrorMsg{Err: err}
			}
		}
		return func() tea.Msg {
			return message.ClipboardCopiedMsg(loadbalancer.PublicNet.IPv6.IP.String())
		}
	case "copy_private_ip":
		// Copy the private IP to clipboard
		if loadbalancer.PrivateNet[0].IP != nil {
			return func() tea.Msg {
				if err := clipboard.WriteAll(loadbalancer.PrivateNet[0].IP.String()); err != nil {
					return message.ErrorMsg{Err: err}
				}
				return message.ClipboardCopiedMsg(loadbalancer.PrivateNet[0].IP.String())
			}
		}
		return func() tea.Msg {
		return message.ErrorMsg{Err: fmt.Errorf("loadbalancer has no private IP")}
		}

	default:
		return nil
	}
}
