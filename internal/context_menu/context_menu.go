package context_menu

import (
	"github.com/grammeaway/lazyhetzner/internal/resource"
)


// Context menu items
type ContextMenuItem struct {
	Label  string
	Action string
}

type ContextMenu struct {
	Items        []ContextMenuItem
	SelectedItem int
	ResourceType resource.ResourceType
	ResourceID   int64
}
