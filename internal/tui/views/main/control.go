package mainview

import (
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
)

const selectionHelpText = "←/→/↑/↓ a/d/w/s h/j/k/l • p: ports • ?: help • q: quit"

const (
	docMarginX       = 1
	cardContentWidth = 20
	cardPaddingX     = 1
	cardBorderX      = 1

	cardTotalWidth      = cardContentWidth + (2 * cardPaddingX) + (2 * cardBorderX)
	viewHorizontalInset = 2 * docMarginX
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionCloseHelp
	ActionOpenHelp
	ActionOpenPorts
	ActionMoveLeft
	ActionMoveRight
	ActionMoveUp
	ActionMoveDown
	ActionStartScan
)

type ScanSelection struct {
	Iface      net.Interface
	Addrs      []net.Addr
	TargetAddr string
	TotalHosts int
}

type UpdateResult struct {
	Model     Model
	Quit      bool
	OpenPorts bool
	StartScan bool
	Selection ScanSelection
}

func HandleKey(showHelp bool, key string) Action {
	if showHelp {
		return ActionCloseHelp
	}

	switch key {
	case "ctrl+c", "q":
		return ActionQuit
	case "?":
		return ActionOpenHelp
	case "p":
		return ActionOpenPorts
	case "left", "a", "h":
		return ActionMoveLeft
	case "right", "d", "l":
		return ActionMoveRight
	case "up", "w", "k":
		return ActionMoveUp
	case "down", "s", "j":
		return ActionMoveDown
	case "enter":
		return ActionStartScan
	default:
		return ActionNone
	}
}

func MoveCursorLeft(cursor int) int {
	if cursor > 0 {
		cursor--
	}
	return cursor
}

func MoveCursorRight(cursor, maxIndex int) int {
	if cursor < maxIndex {
		cursor++
	}
	return cursor
}

func MoveCursorUp(cursor, cardsPerRow int) int {
	next := cursor - cardsPerRow
	if next < 0 {
		return cursor
	}
	return next
}

func MoveCursorDown(cursor, cardsPerRow, maxIndex int) int {
	next := cursor + cardsPerRow
	if next > maxIndex {
		return cursor
	}
	return next
}

func CardsPerRow(windowWidth int) int {
	cardsPerRow := (windowWidth - viewHorizontalInset) / cardTotalWidth
	if cardsPerRow < 1 {
		return 1
	}
	return cardsPerRow
}

func ResolveScanSelection(interfaces []net.Interface, cursor int, addrsByIface map[string][]net.Addr) (ScanSelection, error) {
	if cursor < 0 || cursor >= len(interfaces) {
		return ScanSelection{}, nil
	}

	selection := ScanSelection{Iface: interfaces[cursor]}
	selection.Addrs = addrsByIface[selection.Iface.Name]

	if len(selection.Addrs) == 0 {
		return ScanSelection{}, fmt.Errorf("interface %s has no IP addresses", selection.Iface.Name)
	}

	selection.TargetAddr = scanner.FirstIp4(selection.Addrs)
	if selection.TargetAddr == "" {
		return ScanSelection{}, fmt.Errorf("interface %s has no valid IPv4 addresses", selection.Iface.Name)
	}

	_, ipnet, _ := net.ParseCIDR(selection.TargetAddr)
	selection.TotalHosts = scanner.TotalScanHosts(ipnet)

	return selection, nil
}

func (m Model) Update(msg tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	switch HandleKey(m.ShowHelp, msg.String()) {
	case ActionQuit:
		result.Quit = true
	case ActionCloseHelp:
		result.Model.ShowHelp = false
	case ActionOpenHelp:
		result.Model.ShowHelp = true
	case ActionOpenPorts:
		result.OpenPorts = true
	case ActionMoveLeft:
		result.Model.Cursor = MoveCursorLeft(result.Model.Cursor)
	case ActionMoveRight:
		result.Model.Cursor = MoveCursorRight(result.Model.Cursor, len(result.Model.Interfaces)-1)
	case ActionMoveUp:
		result.Model.Cursor = MoveCursorUp(result.Model.Cursor, result.Model.CardsPerRow)
	case ActionMoveDown:
		result.Model.Cursor = MoveCursorDown(result.Model.Cursor, result.Model.CardsPerRow, len(result.Model.Interfaces)-1)
	case ActionStartScan:
		selection, err := ResolveScanSelection(result.Model.Interfaces, result.Model.Cursor, result.Model.InterfaceMap)
		if err != nil {
			result.Model.ErrorMsg = err.Error()
			return result
		}
		if selection.Iface.Name == "" {
			return result
		}
		result.Model.ErrorMsg = ""
		result.StartScan = true
		result.Selection = selection
	}

	return result
}
