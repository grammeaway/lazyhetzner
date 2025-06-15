package project 




// Project list item
type projectItem struct {
	config    config.ProjectConfig
	isDefault bool
}

func (i projectItem) FilterValue() string { return i.config.Name }
func (i projectItem) Title() string {
	if i.isDefault {
		return fmt.Sprintf("⭐ %s", i.config.Name)
	}
	return i.config.Name
}
func (i projectItem) Description() string {
	tokenPreview := i.config.Token
	if len(tokenPreview) > 16 {
		tokenPreview = tokenPreview[:16] + "..."
	}
	if i.isDefault {
		return fmt.Sprintf("Token: %s (default project)", tokenPreview)
	}
	return fmt.Sprintf("Token: %s", tokenPreview)
}
