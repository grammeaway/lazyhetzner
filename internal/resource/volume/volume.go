package volume

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"lazyhetzner/internal/message"
)

type VolumesLoadedMsg struct {
	Volumes []*hcloud.Volume
}

type VolumeItem struct {
	Volume *hcloud.Volume
}

func (i VolumeItem) FilterValue() string { return i.Volume.Name }
func (i VolumeItem) Title() string       { return i.Volume.Name }
func (i VolumeItem) Description() string {
	status := "📦 Available"
	if i.Volume.Server != nil {
		status = "🔗 Attached to " + i.Volume.Server.Name
	}
	return fmt.Sprintf("%s | %dGB | %s", status, i.Volume.Size, i.Volume.Location.Name)
}

func LoadVolumes(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return VolumesLoadedMsg{Volumes: volumes}
	}
}
