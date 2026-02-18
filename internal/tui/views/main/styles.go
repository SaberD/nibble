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
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, cardPaddingX).
			Width(cardWidth).
			MarginBottom(0)

	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("226")).
				Padding(0, cardPaddingX).
				Width(cardWidth).
				MarginBottom(0)
)
