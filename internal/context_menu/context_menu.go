package context_menu

import (
		"github.com/hetznercloud/hcloud-go/hcloud"
)


// Context menu items
type ContextMenuItem struct {
	Label  string
	Action string
}

type ContextMenu struct {
	Items        []ContextMenuItem
	SelectedItem int
	Server       *hcloud.Server
}
