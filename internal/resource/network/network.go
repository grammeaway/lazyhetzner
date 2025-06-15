package network

import (
	"context"
	"fmt"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/message"
	tea "github.com/charmbracelet/bubbletea"
)

type networksLoadedMsg struct {
	networks []*hcloud.Network
}



type networkItem struct {
	network *hcloud.Network
}

func (i networkItem) FilterValue() string { return i.network.Name }
func (i networkItem) Title() string       { return i.network.Name }
func (i networkItem) Description() string {
	return fmt.Sprintf("IP Range: %s | Subnets: %d", i.network.IPRange.String(), len(i.network.Subnets))
}


func loadNetworks(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		networks, err := client.Network.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return networksLoadedMsg{networks: networks}
	}
}
