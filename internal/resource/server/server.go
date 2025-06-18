package server

import (

	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/message"
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"lazyhetzner/internal/resource"

)


type ServerItem struct {
	Server *hcloud.Server
	ResourceType resource.ResourceType
	ResourceID   int

}

func (i ServerItem) FilterValue() string { return i.Server.Name }
func (i ServerItem) Title() string       { return i.Server.Name }
func (i ServerItem) Description() string {
	var statusDisplay string
	if i.Server.Status == hcloud.ServerStatusRunning {
		statusDisplay = "🟢 " + string(i.Server.Status)
	} else if i.Server.Status == hcloud.ServerStatusStarting || i.Server.Status == hcloud.ServerStatusInitializing {
		statusDisplay = "🟡 " + string(i.Server.Status)
	} else {
		statusDisplay = "🔴 " + string(i.Server.Status)
	}
	return fmt.Sprintf("%s | %s | %s | %s", statusDisplay, i.Server.ServerType.Name, i.Server.PublicNet.IPv4.IP.String(), i.Server.PrivateNet[0].IP.String())

}



func LoadServers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		servers, err := client.Server.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return ServersLoadedMsg{Servers: servers}
	}
}

func CreateServerSnapshot(client *hcloud.Client, server *hcloud.Server, name *string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		createOpts := hcloud.ServerCreateImageOpts{
			Description: name,
			Type: hcloud.ImageTypeSnapshot,
			Labels: map[string]string{
		"created_by": "lazyhetzner",
			},
		}
		snapshot, _, err := client.Server.CreateImage(ctx, server, &createOpts)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return ServerSnapshotCreationStartedMsg {
			Server:        server,
			SnapshotName: snapshot.Image.Description,

		}
	}
}
