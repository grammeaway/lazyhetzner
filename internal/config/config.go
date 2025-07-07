package config

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grammeaway/lazyhetzner/internal/message"
	"os"
	"path/filepath"
)

// Config management
type ProjectConfig struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type Config struct {
	Projects        []ProjectConfig `json:"projects"`
	DefaultProject  string          `json:"default_project"`
	DefaultTerminal string          `json:"default_terminal"`
}

type ConfigLoadedMsg struct {
	Config *Config
}

type ProjectSavedMsg struct{}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configDirComplete := filepath.Join(configDir, "lazyhetzner")
	if err := os.MkdirAll(configDirComplete, os.ModePerm); err != nil {
		return "", err
	}

	return filepath.Join(configDirComplete, "config.json"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{Projects: []ProjectConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Projects: []ProjectConfig{}}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) AddProject(name, token string) {
	// Remove existing project with same name
	for i, p := range c.Projects {
		if p.Name == name {
			c.Projects = append(c.Projects[:i], c.Projects[i+1:]...)
			break
		}
	}

	c.Projects = append(c.Projects, ProjectConfig{
		Name:  name,
		Token: token,
	})

	// Set as default if it's the first project
	if len(c.Projects) == 1 {
		c.DefaultProject = name
	}
}

func (c *Config) GetProject(name string) *ProjectConfig {
	for _, p := range c.Projects {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func (c *Config) RemoveProject(name string) {
	for i, p := range c.Projects {
		if p.Name == name {
			c.Projects = append(c.Projects[:i], c.Projects[i+1:]...)
			if c.DefaultProject == name && len(c.Projects) > 0 {
				c.DefaultProject = c.Projects[0].Name
			}
			break
		}
	}
}

func LoadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		config, err := loadConfig()
		if err != nil {
			return message.ErrorMsg{err}
		}
		return ConfigLoadedMsg{config}
	}
}

func SaveConfigCmd(config *Config) tea.Cmd {
	return func() tea.Msg {
		if err := saveConfig(config); err != nil {
			return message.ErrorMsg{err}
		}
		return ProjectSavedMsg{}
	}
}
