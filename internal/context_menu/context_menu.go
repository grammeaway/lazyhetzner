package context_menu

import (
		"github.com/hetznercloud/hcloud-go/hcloud"
)


// Context menu items
type contextMenuItem struct {
	label  string
	action string
}

type contextMenu struct {
	items        []contextMenuItem
	selectedItem int
	server       *hcloud.Server
}
