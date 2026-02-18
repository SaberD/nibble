package portsview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"

	tea "github.com/charmbracelet/bubbletea"
)

const portsHelpText = "tab • ←/→ a/d h/l • type • backspace: remove • delete: clear all • enter • ?: help • q: quit"

type Action struct {
	Handled    bool
	Quit       bool
	CloseHelp  bool
	OpenHelp   bool
	ToggleMode bool
	Apply      bool
	MoveLeft   bool
	MoveRight  bool
	MoveHome   bool
	MoveEnd    bool
	Backspace  bool
	DeleteAll  bool
}

type Result struct {
	Model Model
	Quit  bool
	Done  bool
}

func HandleKey(showHelp bool, key string) Action {
	if showHelp {
		return Action{Handled: true, CloseHelp: true}
	}

	switch key {
	case "ctrl+c", "q":
		return Action{Handled: true, Quit: true}
	case "?":
		return Action{Handled: true, OpenHelp: true}
	case "tab", "up", "down":
		return Action{Handled: true, ToggleMode: true}
	case "enter":
		return Action{Handled: true, Apply: true}
	case "left", "a", "h":
		return Action{Handled: true, MoveLeft: true}
	case "right", "d", "l":
		return Action{Handled: true, MoveRight: true}
	case "home", "ctrl+a":
		return Action{Handled: true, MoveHome: true}
	case "end", "ctrl+e":
		return Action{Handled: true, MoveEnd: true}
	case "backspace":
		return Action{Handled: true, Backspace: true}
	case "delete":
		return Action{Handled: true, DeleteAll: true}
	default:
		return Action{}
	}
}

func ToggleMode(portPack string) string {
	if portPack == "default" {
		return "custom"
	}
	return "default"
}

func ClampCursor(cursor, valueLen int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > valueLen {
		return valueLen
	}
	return cursor
}

func MoveCursorLeft(cursor int) int {
	if cursor > 0 {
		cursor--
	}
	return cursor
}

func MoveCursorRight(cursor, valueLen int) int {
	if cursor < valueLen {
		cursor++
	}
	return cursor
}

func Backspace(value string, cursor int) (string, int) {
	if cursor <= 0 || len(value) == 0 {
		return value, ClampCursor(cursor, len(value))
	}
	i := cursor - 1
	return value[:i] + value[cursor:], i
}

func InsertRunes(value string, cursor int, runes []rune) (string, int) {
	cursor = ClampCursor(cursor, len(value))
	for _, r := range runes {
		if r >= '0' && r <= '9' {
			if !canInsertPortDigit(value, cursor, r) {
				continue
			}
			s := string(r)
			value = value[:cursor] + s + value[cursor:]
			cursor++
			continue
		}
		if r == ',' {
			s := string(r)
			value = value[:cursor] + s + value[cursor:]
			cursor++
		}
	}
	return value, cursor
}

func canInsertPortDigit(s string, cursor int, digit rune) bool {
	start, end := currentTokenBounds(s, cursor)
	pos := cursor - start
	token := s[start:end]
	next := token[:pos] + string(digit) + token[pos:]
	return len(next) <= 5
}

func currentTokenBounds(s string, cursor int) (int, int) {
	cursor = ClampCursor(cursor, len(s))
	start := -1
	for i := cursor - 1; i >= 0; i-- {
		if s[i] == ',' {
			start = i
			break
		}
	}
	if start == -1 {
		start = 0
	} else {
		start++
	}
	end := len(s)
	for i := cursor; i < len(s); i++ {
		if s[i] == ',' {
			end = i
			break
		}
	}
	return start, end
}

func applyConfig(m Model) (Model, bool) {
	addPorts := ""
	if m.PortPack == "custom" {
		normalized, err := ports.NormalizeCustom(strings.TrimSpace(m.CustomPorts))
		if err != nil {
			m.ErrorMsg = err.Error()
			return m, false
		}
		addPorts = normalized
		m.CustomPorts = normalized
		m.CustomCursor = len(normalized)
	}

	resolvedPorts, err := ports.Resolve(m.PortPack, addPorts, "")
	if err != nil {
		m.ErrorMsg = err.Error()
		return m, false
	}
	if err := ports.SaveConfig(ports.Config{Mode: m.PortPack, Custom: addPorts}); err != nil {
		m.ErrorMsg = err.Error()
		return m, false
	}
	if netScanner, ok := m.NetworkScan.(*scan.NetScanner); ok {
		netScanner.Ports = resolvedPorts
	}
	m.ErrorMsg = ""
	return m, true
}

func (m Model) Update(msg tea.KeyMsg) Result {
	result := Result{Model: m}
	action := HandleKey(m.ShowHelp, msg.String())
	if action.Quit {
		result.Quit = true
		return result
	}
	if action.CloseHelp {
		result.Model.ShowHelp = false
		return result
	}
	if action.OpenHelp {
		result.Model.ShowHelp = true
		return result
	}
	if action.ToggleMode {
		result.Model.PortPack = ToggleMode(result.Model.PortPack)
		result.Model.CustomCursor = ClampCursor(result.Model.CustomCursor, len(result.Model.CustomPorts))
		return result
	}
	if action.Apply {
		next, ok := applyConfig(result.Model)
		result.Model = next
		result.Done = ok
		return result
	}
	if action.MoveLeft {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor > 0 {
			result.Model.CustomCursor = MoveCursorLeft(result.Model.CustomCursor)
		}
		return result
	}
	if action.MoveRight {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor < len(result.Model.CustomPorts) {
			result.Model.CustomCursor = MoveCursorRight(result.Model.CustomCursor, len(result.Model.CustomPorts))
		}
		return result
	}
	if action.MoveHome {
		if result.Model.PortPack == "custom" {
			result.Model.CustomCursor = 0
		}
		return result
	}
	if action.MoveEnd {
		if result.Model.PortPack == "custom" {
			result.Model.CustomCursor = len(result.Model.CustomPorts)
		}
		return result
	}
	if action.Backspace {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor > 0 && len(result.Model.CustomPorts) > 0 {
			result.Model.CustomPorts, result.Model.CustomCursor = Backspace(result.Model.CustomPorts, result.Model.CustomCursor)
		}
		return result
	}
	if action.DeleteAll {
		if result.Model.PortPack == "custom" {
			result.Model.CustomPorts = ""
			result.Model.CustomCursor = 0
		}
		return result
	}
	if result.Model.PortPack == "custom" && msg.Type == tea.KeyRunes {
		result.Model.CustomPorts, result.Model.CustomCursor = InsertRunes(result.Model.CustomPorts, result.Model.CustomCursor, msg.Runes)
	}
	return result
}
