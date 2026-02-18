package scan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	bannerReadTimeout = 300 * time.Millisecond
	maxBannerLength   = 80
)

// getServiceBanner reads a service banner
// Prefer passive reads first, then fall back to HTTP probe.
func getServiceBanner(conn net.Conn) string {
	_ = conn.SetDeadline(time.Now().Add(bannerReadTimeout))

	passive := readPushBanner(conn)
	if passive != "" {
		if checkIfHttp(passive) {
			return parseHTTPServer(passive)
		}
		return passive
	}

	return requestHttpBanner(conn)
}

// requestHttpBanner sends a simple HEAD request and returns a parsed banner string
func requestHttpBanner(conn net.Conn) string {
	_, _ = fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: %s\r\n\r\n", conn.RemoteAddr())

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return ""
	}

	response := string(buf[:n])
	if checkIfHttp(response) {
		return parseHTTPServer(response)
	}
	return cleanBanner(buf[:n])
}

func readPushBanner(conn net.Conn) string {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return ""
	}

	return cleanBanner(buf[:n])
}

// cleanBanner normalizes raw banner bytes into a short printable string
func cleanBanner(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	// Keep only the first line
	if idx := firstLineBreak(raw); idx >= 0 {
		raw = raw[:idx]
	}

	s := replaceInvalid(raw)

	// Keep printable ASCII only, and trim whitespace
	out := []rune{}
	for _, r := range strings.TrimSpace(s) {
		if r >= 32 && r <= 126 {
			out = append(out, r)
		}
	}

	// Enforce maximum banner length
	clean := strings.TrimSpace(string(out))
	if len(clean) <= maxBannerLength {
		return clean
	}
	return clean[:maxBannerLength]
}

// replaceInvalid replaces invalid UTF-8 bytes with '.'
func replaceInvalid(raw []byte) string {
	s := string(raw)
	if utf8.ValidString(s) {
		return s
	}

	cleaned := make([]rune, 0, len(raw))
	for len(raw) > 0 {
		r, size := utf8.DecodeRune(raw)
		if r == utf8.RuneError && size == 1 {
			cleaned = append(cleaned, '.')
			raw = raw[1:]
			continue
		}
		cleaned = append(cleaned, r)
		raw = raw[size:]
	}

	return string(cleaned)
}

func firstLineBreak(b []byte) int {
	newlineIndex := newlineByte(b)
	returnIndex := returnByte(b)

	switch {
	case newlineIndex < 0:
		return returnIndex
	case returnIndex < 0:
		return newlineIndex
	case newlineIndex < returnIndex:
		return newlineIndex
	default:
		return returnIndex
	}
}

// newlineByte finds first newline byte "\n"
func newlineByte(b []byte) int {
	for i, c := range b {
		if c == '\n' {
			return i
		}
	}
	return -1
}

// returnByte finds first return byte "\r"
func returnByte(b []byte) int {
	for i, c := range b {
		if c == '\r' {
			return i
		}
	}
	return -1
}

// parseHTTPServer extracts the Server header from an HTTP response
func parseHTTPServer(response string) string {
	for _, line := range strings.Split(response, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "server:") {
			return cleanBanner([]byte(strings.TrimSpace(line[7:])))
		}
	}

	if idx := strings.Index(response, "\r\n"); idx > 0 {
		return cleanBanner([]byte(response[:idx]))
	}
	return cleanBanner([]byte(response))
}

func checkIfHttp(s string) bool {
	line := s
	if idx := strings.IndexAny(s, "\r\n"); idx >= 0 {
		line = s[:idx]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return false
	}
	if !strings.HasPrefix(strings.ToUpper(parts[0]), "HTTP/") {
		return false
	}

	code, err := strconv.Atoi(parts[1])
	return err == nil && code >= 100 && code <= 599
}
