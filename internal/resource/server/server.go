package server

import (

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/resource"

)


type ServersLoadedMsg struct {
	Servers []*hcloud.Server
}



type ServerItem struct {
	Server *hcloud.Server
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
