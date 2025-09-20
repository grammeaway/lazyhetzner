package volume

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

func CreateVolumeContextMenu(volume *hcloud.Volume) ctm.ContextMenu {
	return ctm.ContextMenu{
		Items:        getVolumeMenuItems(),
		SelectedItem: 0,
		ResourceType: resource.ResourceVolumes,
		ResourceID:   volume.ID,
	}
}

func getVolumeLabels(volume *hcloud.Volume) map[string]string {
	labels := make(map[string]string)
	for key, value := range volume.Labels {
		labels[key] = value
	}
	return labels
}

// Returns context menu Items for volumes
func getVolumeMenuItems() []ctm.ContextMenuItem {
	return []ctm.ContextMenuItem{
		// add action for canceling (i.e., closing) the context menu
		{Label: "❌ Cancel", Action: "cancel"},
		{Label: "🔖 View Labels", Action: "view_labels"},
		// Copy the volume ID to clipboard
		{Label: "📋 Copy Volume ID", Action: "copy_id"},
		// Copy the volume name to clipboard
		{Label: "📋 Copy Volume Name", Action: "copy_name"},
		// Copy the attached server ID to clipboard 
		{Label: "📋 Copy Attached Server ID", Action: "copy_server_id"},
		// Copy the attached server name to clipboard
		{Label: "📋 Copy Attached Server Name", Action: "copy_server_name"},
	}
}

func ExecuteVolumeContextAction(selectedAction string, volume *hcloud.Volume) tea.Cmd {
	switch selectedAction {
	case "cancel":
		return func() tea.Msg {
			return message.CancelCtxMenuMsg{}
		}
	case "view_labels":
		labels := getVolumeLabels(volume)
		if len(labels) == 0 {
			return func() tea.Msg {
				return message.StatusMsg("No labels found for this volume.")
			}
		}
		return func() tea.Msg {
			return label.LabelsLoadedMsg{
				Labels:              labels,
				RelatedResourceName: fmt.Sprintf("Volume: %s", volume.Name),
				RelatedResourceType: resource.ResourceServers,
			}
		}
	case "copy_id":
		// Copy the volume ID to clipboard
		return func() tea.Msg {
			if err := clipboard.WriteAll(fmt.Sprintf("%d", volume.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Volume ID %d copied to clipboard", volume.ID))
		}
	case "copy_name":
		// Copy the volume name to clipboard
		return func() tea.Msg {
			if volume.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("Volume has no name")}
			}
			if err := clipboard.WriteAll(volume.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Volume name '%s' copied to clipboard", volume.Name))
		}
	case "copy_server_id":
		// Copy the attached server ID to clipboard
		return func() tea.Msg {
			if volume.Server == nil {
				return message.ErrorMsg{Err: fmt.Errorf("Volume is not attached to any server")}
			}
			if err := clipboard.WriteAll(fmt.Sprintf("%d", volume.Server.ID)); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Attached Server ID %d copied to clipboard", volume.Server.ID))
		}
	case "copy_server_name":
		// Copy the attached server name to clipboard
		return func() tea.Msg {
			if volume.Server == nil {
				return message.ErrorMsg{Err: fmt.Errorf("Volume is not attached to any server")}
			}
			if volume.Server.Name == "" {
				return message.ErrorMsg{Err: fmt.Errorf("Attached server has no name")}
			}
			if err := clipboard.WriteAll(volume.Server.Name); err != nil {
				return message.ErrorMsg{Err: err}
			}
			return message.ClipboardCopiedMsg(fmt.Sprintf("Attached Server name '%s' copied to clipboard", volume.Server.Name))
		}

	default:
		return nil
	}
}
