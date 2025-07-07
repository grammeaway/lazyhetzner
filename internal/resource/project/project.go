package project 


import (
	"fmt"
	"github.com/grammeaway/lazyhetzner/internal/config"
)

// Project list item
type ProjectItem struct {
	Config    config.ProjectConfig
	IsDefault bool
}

func (i ProjectItem) FilterValue() string { return i.Config.Name }
func (i ProjectItem) Title() string {
	if i.IsDefault {
		return fmt.Sprintf("⭐ %s", i.Config.Name)
	}
	return i.Config.Name
}
func (i ProjectItem) Description() string {
	tokenPreview := i.Config.Token
	if len(tokenPreview) > 16 {
		tokenPreview = tokenPreview[:16] + "..."
	}
	if i.IsDefault {
		return fmt.Sprintf("Token: %s (default project)", tokenPreview)
	}
	return fmt.Sprintf("Token: %s", tokenPreview)
}
