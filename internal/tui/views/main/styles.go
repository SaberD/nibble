package mainview

import "github.com/charmbracelet/lipgloss"

const (
	cardWidth    = 20
	cardPaddingX = 1

	cardTotalWidth = cardWidth + 2*cardPaddingX
)

func CardsPerRow(windowWidth int) int {
	cardsPerRow := windowWidth / cardTotalWidth
	if cardsPerRow < 1 {
		return 1
	}
	return cardsPerRow
}

var (
	baseCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, cardPaddingX).
			Width(cardWidth).
			MarginBottom(0)

	cardStyle         = baseCardStyle.BorderForeground(lipgloss.Color("8"))
	selectedCardStyle = baseCardStyle.BorderForeground(lipgloss.Color("226"))
)
