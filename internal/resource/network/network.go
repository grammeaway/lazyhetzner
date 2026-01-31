package network

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type NetworksLoadedMsg struct {
	Networks []*hcloud.Network
}

type ViewNetworkSubnetsMsg struct {
	Network *hcloud.Network
	Subnets []hcloud.NetworkSubnet
}

type NetworkItem struct {
	Network *hcloud.Network

	ResourceType resource.ResourceType
	ResourceID   int64
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
