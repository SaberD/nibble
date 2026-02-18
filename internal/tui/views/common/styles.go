package common

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	DocStyle = lipgloss.NewStyle().
			Margin(0, 1)
)
