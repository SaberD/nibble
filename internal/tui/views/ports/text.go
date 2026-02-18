package portsview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// wrapPortList wraps comma-separated content window max width
func wrapPortList(prefix, content string, maxWidth int) string {
	if content == "" {
		return prefix
	}
	if maxWidth <= len(prefix)+1 {
		return prefix + content
	}

	indent := strings.Repeat(" ", len(prefix))
	tokens := strings.Split(content, ",")
	lines := make([]string, 0, 4)
	current := prefix

	for i, token := range tokens {
		segment := token
		if i < len(tokens)-1 {
			segment += ","
		}
		if len(current)+len(segment) > maxWidth && len(current) > len(prefix) {
			lines = append(lines, current)
			current = indent + segment
			continue
		}
		current += segment
	}

	lines = append(lines, current)
	return strings.Join(lines, "\n")
}

func invalidPorts(errMsg string) []string {
	const prefix = "invalid ports: "
	lower := strings.ToLower(errMsg)
	if !strings.Contains(lower, prefix) {
		return nil
	}
	i := strings.Index(lower, prefix)
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(errMsg[i+len(prefix):])
	if rest == "" {
		return nil
	}
	parts := strings.Split(rest, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func highlightInvalidPorts(s string, tokens []string) string {
	invalidStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	for _, token := range tokens {
		if token == "" {
			continue
		}
		start := 0
		for {
			idx := strings.Index(s[start:], token)
			if idx < 0 {
				break
			}
			idx += start
			end := idx + len(token)

			prevOK := idx == 0 || strings.ContainsRune(" ,:|\n", rune(s[idx-1]))
			nextOK := end == len(s) || strings.ContainsRune(",|\n ", rune(s[end]))
			if prevOK && nextOK {
				s = s[:idx] + invalidStyle.Render(token) + s[end:]
				break
			}
			start = end
		}
	}
	return s
}
