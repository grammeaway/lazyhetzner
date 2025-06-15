package model

import (
	"github.com/charmbracelet/lipgloss"
)
// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a1a"))

	selectedMenuStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#874BFD")).
				Foreground(lipgloss.Color("#FFFDF5")).
				Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	focusedStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#874BFD"))

	blurredStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#626262"))
)

