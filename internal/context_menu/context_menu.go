package context_menu

import (
		"github.com/hetznercloud/hcloud-go/hcloud"
	"lazyhetzner/internal/resource"
)


// Context menu items
type ContextMenuItem struct {
	Label  string
	Action string
	ResourceType resource.ResourceType
	ResourceID   int
}

type ContextMenu struct {
	Items        []ContextMenuItem
	SelectedItem int
	Server       *hcloud.Server
}
