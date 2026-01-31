package floatingip

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"github.com/grammeaway/lazyhetzner/internal/resource"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type FloatingIPsLoadedMsg struct {
	FloatingIPs []*hcloud.FloatingIP
}

type FloatingIPItem struct {
	FloatingIP   *hcloud.FloatingIP
	ResourceType resource.ResourceType
	ResourceID   int64
}

func (i FloatingIPItem) FilterValue() string { return floatingIPDisplayName(i.FloatingIP) }
func (i FloatingIPItem) Title() string       { return floatingIPDisplayName(i.FloatingIP) }
func (i FloatingIPItem) Description() string {
	status := "🟢 Unassigned"
	if i.FloatingIP.Blocked {
		status = "⛔ Blocked"
	} else if i.FloatingIP.Server != nil {
		status = "🔗 Attached to " + i.FloatingIP.Server.Name
	}
	ipAddress := "N/A"
	if i.FloatingIP.IP != nil {
		ipAddress = i.FloatingIP.IP.String()
	}
	protocol := floatingIPProtocolDisplay(i.FloatingIP.Type)
	return fmt.Sprintf("%s | %s | %s", status, protocol, ipAddress)
}

func LoadFloatingIPs(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		floatingIPs, err := client.FloatingIP.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		for _, floatingIP := range floatingIPs {
			if floatingIP.Server != nil {
				server, _, err := client.Server.GetByID(ctx, floatingIP.Server.ID)
				if err == nil && server != nil {
					floatingIP.Server = server
				}
			}
		}

		return FloatingIPsLoadedMsg{FloatingIPs: floatingIPs}
	}
}

func floatingIPDisplayName(floatingIP *hcloud.FloatingIP) string {
	if floatingIP == nil {
		return "Unknown Floating IP"
	}
	name := strings.TrimSpace(floatingIP.Name)
	if name != "" {
		return name
	}
	if floatingIP.IP != nil {
		return floatingIP.IP.String()
	}
	return fmt.Sprintf("Floating IP %d", floatingIP.ID)
}

func floatingIPProtocolDisplay(protocol hcloud.FloatingIPType) string {
	switch protocol {
	case hcloud.FloatingIPTypeIPv4:
		return "🧭 IPv4"
	case hcloud.FloatingIPTypeIPv6:
		return "🌐 IPv6"
	default:
		return fmt.Sprintf("❓ %s", strings.ToUpper(string(protocol)))
	}
}
