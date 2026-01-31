package firewall

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type FirewallsLoadedMsg struct {
	Firewalls []*hcloud.Firewall
}

type ViewFirewallRulesMsg struct {
	Firewall *hcloud.Firewall
	Rules    []hcloud.FirewallRule
}

type FirewallItem struct {
	Firewall *hcloud.Firewall

	ResourceType resource.ResourceType
	ResourceID   int64
}

func (i FirewallItem) FilterValue() string { return i.Firewall.Name }
func (i FirewallItem) Title() string       { return i.Firewall.Name }
func (i FirewallItem) Description() string {
	return fmt.Sprintf("Rules: %d | Applied to: %d", len(i.Firewall.Rules), len(i.Firewall.AppliedTo))
}

func LoadFirewalls(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		firewalls, err := client.Firewall.All(ctx)
		if err != nil {
			return message.ErrorMsg{Err: err}
		}
		return FirewallsLoadedMsg{Firewalls: firewalls}
	}
}
