package network

import (
	"context"
	"fmt"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/message"
	tea "github.com/charmbracelet/bubbletea"
	"lazyhetzner/internal/resource"
)

type NetworksLoadedMsg struct {
	Networks []*hcloud.Network
}



type NetworkItem struct {
	Network *hcloud.Network

	ResourceType resource.ResourceType
	ResourceID   int
}

func (i NetworkItem) FilterValue() string { return i.Network.Name }
func (i NetworkItem) Title() string       { return i.Network.Name }
func (i NetworkItem) Description() string {
	return fmt.Sprintf("IP Range: %s | Subnets: %d", i.Network.IPRange.String(), len(i.Network.Subnets))
}


func LoadNetworks(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		networks, err := client.Network.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return NetworksLoadedMsg{Networks: networks}
	}
}
