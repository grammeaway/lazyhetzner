package volume

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/message"
)

type volumesLoadedMsg struct {
	volumes []*hcloud.Volume
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

func loadVolumes(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return volumesLoadedMsg{volumes: volumes}
	}
}
