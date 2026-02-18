package mainview

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	cardsPerRow := m.CardsPerRow
	if cardsPerRow == 0 {
		cardsPerRow = 1
	}
	var b strings.Builder

	titleText := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Render("Nibble Network Scanner")
	b.WriteString(titleText + "\n")

	icons := make(map[string]string, len(m.Interfaces))
	for _, iface := range m.Interfaces {
		icons[iface.Name] = interfaceIcon(iface.Name)
	}

	var rows []string
	var currentRow []string
	for i, iface := range m.Interfaces {
		card := renderInterfaceCard(m, icons, i, iface)
		currentRow = append(currentRow, card)
		if len(currentRow) == cardsPerRow || i == len(m.Interfaces)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}

	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, rows...))
	view := b.String()

	if m.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		view += "\n\n" + errorStyle.Render("Error: "+m.ErrorMsg)
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	view += "\n" + helpStyle.Render(common.WrapWords(selectionHelpText, maxWidth))

	if m.ShowHelp {
		return renderHelpOverlay(view)
	}
	return view
}

func renderInterfaceCard(m Model, icons map[string]string, index int, iface net.Interface) string {
	isSelected := index == m.Cursor
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	var cardContent strings.Builder
	name := iface.Name
	icon := icons[name]
	if icon == "" {
		icon = "🔌"
	}

	nameStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(lipgloss.Color("226"))
	}
	cardContent.WriteString(nameStyle.Render(icon+" "+name) + "\n")

	addrs := ipv4Labels(m.InterfaceMap, name)
	addrStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if len(addrs) > 0 {
		cardContent.WriteString(addrStyle.Render(addrs[0]))
	}

	return style.Render(cardContent.String())
}

// ipv4Labels returns IPv4 labels for an interface
func ipv4Labels(addrsByIface map[string][]net.Addr, name string) []string {
	labels := make([]string, 0)
	for _, addr := range addrsByIface[name] {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ones, _ := ipnet.Mask.Size()
			labels = append(labels, fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
		}
	}
	return labels
}
