package server

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type ServersLoadedMsg struct {
	Servers []*hcloud.Server
}

type ServerDetailsLoadedMsg struct {
	Server   *hcloud.Server
	Networks []*hcloud.Network
}

type ServerItem struct {
	Server       *hcloud.Server
	ResourceType resource.ResourceType
	ResourceID   int64
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
	privateIP := "N/A"
	if len(i.Server.PrivateNet) > 0 {
		privateIP = i.Server.PrivateNet[0].IP.String()
	}
	publicIP := "N/A"
	if i.Server.PublicNet.IPv4.IP != nil {
		publicIP = i.Server.PublicNet.IPv4.IP.String()
	}
	return fmt.Sprintf("%s | %s | %s | %s", statusDisplay, i.Server.ServerType.Name, publicIP, privateIP)
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

func LoadServerDetails(client *hcloud.Client, serverID int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		server, _, err := client.Server.GetByID(ctx, serverID)
		if err != nil {
			return message.ErrorMsg{err}
		}
		if server == nil {
			return message.ErrorMsg{fmt.Errorf("server with ID %d not found", serverID)}
		}

		networks := make([]*hcloud.Network, 0, len(server.PrivateNet))
		for _, privateNet := range server.PrivateNet {
			if privateNet.Network == nil {
				continue
			}
			network, _, err := client.Network.GetByID(ctx, privateNet.Network.ID)
			if err != nil {
				return message.ErrorMsg{err}
			}
			if network != nil {
				networks = append(networks, network)
			}
		}

		return ServerDetailsLoadedMsg{Server: server, Networks: networks}
	}
}
