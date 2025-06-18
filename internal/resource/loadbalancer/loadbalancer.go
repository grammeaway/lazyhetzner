package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/message"
	tea "github.com/charmbracelet/bubbletea"

	"lazyhetzner/internal/resource"
	)

type LoadBalancersLoadedMsg struct {
	LoadBalancers []*hcloud.LoadBalancer
}




type LoadBalancerItem struct {
	Lb *hcloud.LoadBalancer

	ResourceType resource.ResourceType
	ResourceID   int
}

func (i LoadBalancerItem) FilterValue() string { return i.Lb.Name }
func (i LoadBalancerItem) Title() string       { return i.Lb.Name }
func (i LoadBalancerItem) Description() string {
	status := "🟢 Available"
	if i.Lb.PublicNet.Enabled {
		return fmt.Sprintf("%s | %s | Targets: %d", status, i.Lb.PublicNet.IPv4.IP.String(), len(i.Lb.Targets))
	}
	return fmt.Sprintf("%s | Private only | Targets: %d", status, len(i.Lb.Targets))
}



func LoadLoadBalancers(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}
		return LoadBalancersLoadedMsg{LoadBalancers: loadBalancers}
	}
}
