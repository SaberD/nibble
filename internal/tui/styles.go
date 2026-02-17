package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles for the Nibble TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true)

	docStyle = lipgloss.NewStyle().
			Margin(0, 1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Width(22).
			MarginBottom(0)

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("226")).
				Padding(0, 1).
				Width(22).
				MarginBottom(0)
)
