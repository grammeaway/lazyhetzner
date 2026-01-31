package firewall

import (
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	ctm "github.com/grammeaway/lazyhetzner/internal/context_menu"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	r_firewall "github.com/grammeaway/lazyhetzner/internal/resource/firewall"
	"github.com/grammeaway/lazyhetzner/internal/resource/label"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func CreateFirewallContextMenu(firewall *hcloud.Firewall) ctm.ContextMenu {
	return ctm.ContextMenu{
		Items:        getFirewallMenuItems(),
		SelectedItem: 0,
		ResourceType: resource.ResourceFirewalls,
		ResourceID:   firewall.ID,
	}
}

func getFirewallLabels(firewall *hcloud.Firewall) map[string]string {
	labels := make(map[string]string)
	for key, value := range firewall.Labels {
		labels[key] = value
	}
	return labels
}

func getFirewallMenuItems() []ctm.ContextMenuItem {
	return []ctm.ContextMenuItem{
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🔒 View Rules", Action: "view_rules"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		{Label: "📋 Copy Firewall ID", Action: "copy_id"},
		{Label: "📋 Copy Firewall Name", Action: "copy_name"},
	}
}

func ExecuteFirewallContextAction(selectedAction string, firewall *hcloud.Firewall) tea.Cmd {
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_rules":
		return func() tea.Msg {
			if len(firewall.Rules) == 0 {
				return message.StatusMsg("No rules found for this firewall.")
			}
			return r_firewall.ViewFirewallRulesMsg{
				Firewall: firewall,
				Rules:    firewall.Rules,
			}
		}
	case "view_labels":
		labels := getFirewallLabels(firewall)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this firewall.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Firewall: %s", firewall.Name),
				RelatedResourceType: resource.ResourceFirewalls,
			}
		}
	case "copy_id":
		return func() tea.Msg {
			if err := clipboard.WriteAll(fmt.Sprintf("%d", firewall.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("%d", firewall.ID))
		}
	case "copy_name":
		return func() tea.Msg {
			if firewall.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("firewall has no name")}
			}
			if err := clipboard.WriteAll(firewall.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(firewall.Name)
		}
	default:
		return nil
	}
}
