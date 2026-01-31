package model

import (
	"github.com/charmbracelet/lipgloss"
)

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

	labelKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Background(lipgloss.Color("#2a2a2a")).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD"))

	labelValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262"))

	labelContainerStyle = lipgloss.NewStyle().
				Margin(0, 0, 1, 0).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Background(lipgloss.Color("#0f0f0f"))

	noLabelsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Background(lipgloss.Color("#2a1a00")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFAA00")).
			Italic(true)

	targetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262"))

	noTargetsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Background(lipgloss.Color("#2a1a00")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFAA00")).
			Italic(true)

	targetContainerStyle = lipgloss.NewStyle().
				Margin(0, 0, 1, 0).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Background(lipgloss.Color("#0f0f0f"))

	serviceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262"))
	noServicesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Background(lipgloss.Color("#2a1a00")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFAA00")).
			Italic(true)
	serviceContainerStyle = lipgloss.NewStyle().
				Margin(0, 0, 1, 0).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Background(lipgloss.Color("#0f0f0f"))
	firewallRuleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#626262"))
	noFirewallRulesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAA00")).
				Background(lipgloss.Color("#2a1a00")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FFAA00")).
				Italic(true)
	firewallRuleContainerStyle = lipgloss.NewStyle().
					Margin(0, 0, 1, 0).
					Padding(1).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#444444")).
					Background(lipgloss.Color("#0f0f0f"))
	subnetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262"))
	noSubnetsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Background(lipgloss.Color("#2a1a00")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFAA00")).
			Italic(true)
	subnetContainerStyle = lipgloss.NewStyle().
				Margin(0, 0, 1, 0).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Background(lipgloss.Color("#0f0f0f"))
	serverDetailSectionStyle = lipgloss.NewStyle().
					Margin(0).
					Padding(0, 1).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#3b82f6")).
					Background(lipgloss.Color("#0b1a2b"))
	serverDetailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#93c5fd")).
				Bold(true)
)
