package common

import "strings"

func WrapWords(s string, maxWidth int) string {
	if maxWidth <= 0 || len(s) <= maxWidth {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	lines := []string{words[0]}
	for _, w := range words[1:] {
		last := lines[len(lines)-1]
		if len(last)+1+len(w) <= maxWidth {
			lines[len(lines)-1] = last + " " + w
			continue
		}
		lines = append(lines, w)
	}
	return strings.Join(lines, "\n")
}
