package ports

import (
	"fmt"
	"sort"
	"strings"
)

const (
	ModeDefault = "default"
	ModeCustom  = "custom"
)

var defaultPorts = []int{
	22,   // SSH
	23,   // Telnet
	53,   // DNS
	80,   // HTTP
	139,  // NetBIOS Session Service
	443,  // HTTPS
	445,  // SMB
	1883, // MQTT
	3306, // MySQL
	3389, // RDP
	5432, // PostgreSQL
	8000, // Alt HTTP
	8080, // Alt HTTP proxy/app
	8443, // Alt HTTPS
}

func IsValidPack(name string) bool {
	return name == ModeDefault || name == ModeCustom
}

func DefaultPorts() []int {
	out := make([]int, len(defaultPorts))
	copy(out, defaultPorts)
	return out
}

// Resolve returns the final port list from a named pack plus optional add/remove lists.
func Resolve(packName, addPorts, removePorts string) ([]int, error) {
	if packName == "" {
		packName = ModeDefault
	}

	var base []int
	switch packName {
	case ModeDefault:
		base = defaultPorts
	case ModeCustom:
		base = nil
	default:
		return nil, fmt.Errorf("unknown port pack: %s", packName)
	}

	add, err := parseList(addPorts)
	if err != nil {
		return nil, err
	}
	remove, err := parseList(removePorts)
	if err != nil {
		return nil, err
	}

	set := make(map[int]struct{}, len(base)+len(add))
	for _, p := range base {
		set[p] = struct{}{}
	}
	for _, p := range add {
		set[p] = struct{}{}
	}
	for _, p := range remove {
		delete(set, p)
	}

	out := []int{}
	for p := range set {
		out = append(out, p)
	}
	sort.Ints(out)
	return out, nil
}

func parseList(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	fields := strings.Split(raw, ",")
	out := make([]int, 0, len(fields))
	invalid := make([]string, 0, 2)
	for _, f := range fields {
		s := strings.TrimSpace(f)
		if s == "" {
			continue
		}
		ports, err := parseToken(s)
		if err != nil {
			invalid = append(invalid, s)
			continue
		}
		out = append(out, ports...)
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid ports: %s", strings.Join(invalid, ","))
	}

	set := make(map[int]struct{}, len(out))
	for _, p := range out {
		set[p] = struct{}{}
	}

	normalized := make([]int, 0, len(set))
	for p := range set {
		normalized = append(normalized, p)
	}
	sort.Ints(normalized)
	return normalized, nil
}

func parseToken(raw string) ([]int, error) {
	start, end, err := parseTokenBounds(raw)
	if err != nil {
		return nil, err
	}

	out := make([]int, 0, end-start+1)
	for p := start; p <= end; p++ {
		out = append(out, p)
	}
	return out, nil
}

// NormalizeCustom returns a normalized CSV list for custom ports:
// valid tokens only, duplicates removed, and range tokens preserved.
func NormalizeCustom(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}

	fields := strings.Split(raw, ",")
	ranges := make([]portRange, 0, len(fields))
	invalid := make([]string, 0, 2)

	for _, f := range fields {
		s := strings.TrimSpace(f)
		if s == "" {
			continue
		}

		start, end, err := parseTokenBounds(s)
		if err != nil {
			invalid = append(invalid, s)
			continue
		}
		ranges = append(ranges, portRange{start: start, end: end})
	}

	if len(invalid) > 0 {
		return "", fmt.Errorf("invalid ports: %s", strings.Join(invalid, ","))
	}
	return normalizeRanges(ranges), nil
}
