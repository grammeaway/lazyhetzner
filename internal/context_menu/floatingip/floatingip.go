package floatingip

import (
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	ctm "github.com/grammeaway/lazyhetzner/internal/context_menu"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/grammeaway/lazyhetzner/internal/resource/label"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func CreateFloatingIPContextMenu(floatingIP *hcloud.FloatingIP) ctm.ContextMenu {
	return ctm.ContextMenu{
		Items:        getFloatingIPMenuItems(),
		SelectedItem: 0,
		ResourceType: resource.ResourceFloatingIPs,
		ResourceID:   floatingIP.ID,
	}
}

func getFloatingIPLabels(floatingIP *hcloud.FloatingIP) map[string]string {
	labels := make(map[string]string)
	for key, value := range floatingIP.Labels {
		labels[key] = value
	}
	return labels
}

func getFloatingIPMenuItems() []ctm.ContextMenuItem {
	return []ctm.ContextMenuItem{
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		{Label: "📋 Copy Floating IP ID", Action: "copy_id"},
		{Label: "📋 Copy Floating IP Name", Action: "copy_name"},
		{Label: "📋 Copy Floating IP Address", Action: "copy_ip"},
	}
}

func ExecuteFloatingIPContextAction(selectedAction string, floatingIP *hcloud.FloatingIP) tea.Cmd {
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_labels":
		labels := getFloatingIPLabels(floatingIP)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this floating IP.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Floating IP: %s", floatingIPDisplayName(floatingIP)),
				RelatedResourceType: resource.ResourceFloatingIPs,
			}
		}
	case "copy_id":
		return func() tea.Msg {
			if err := clipboard.WriteAll(fmt.Sprintf("%d", floatingIP.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Floating IP ID %d copied to clipboard", floatingIP.ID))
		}
	case "copy_name":
		return func() tea.Msg {
			if floatingIP.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("floating IP has no name")}
			}
			if err := clipboard.WriteAll(floatingIP.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Floating IP name '%s' copied to clipboard", floatingIP.Name))
		}
	case "copy_ip":
		return func() tea.Msg {
			if floatingIP.IP == nil {
				return message.ErrorMsg{Err: fmt.Errorf("floating IP has no address")}
			}
			if err := clipboard.WriteAll(floatingIP.IP.String()); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Floating IP address '%s' copied to clipboard", floatingIP.IP.String()))
		}
	default:
		return nil
	}
}

func floatingIPDisplayName(floatingIP *hcloud.FloatingIP) string {
	if floatingIP == nil {
		return "Unknown Floating IP"
	}
	if floatingIP.Name != "" {
		return floatingIP.Name
	}
	if floatingIP.IP != nil {
		return floatingIP.IP.String()
	}
	return fmt.Sprintf("Floating IP %d", floatingIP.ID)
}
