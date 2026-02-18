package mainview

import "github.com/charmbracelet/lipgloss"

var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Width(20).
			MarginBottom(0)

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("226")).
				Padding(0, 1).
				Width(20).
				MarginBottom(0)
)
