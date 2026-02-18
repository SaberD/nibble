package ports

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var defaultPacks = map[string][]int{
	"default": {22, 23, 80, 443, 445, 3389, 8080},
	"custom":  {},
}

func IsValidPack(name string) bool {
	_, ok := defaultPacks[name]
	return ok
}

func DefaultPorts() []int {
	out := make([]int, len(defaultPacks["default"]))
	copy(out, defaultPacks["default"])
	return out
}

// Resolve returns the final port list from a named pack plus optional add/remove lists.
func Resolve(packName, addPorts, removePorts string) ([]int, error) {
	if packName == "" {
		packName = "default"
	}

	base, ok := defaultPacks[packName]
	if !ok {
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

	out := make([]int, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Ints(out)
	if len(out) == 0 {
		return nil, fmt.Errorf("port list is empty")
	}
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
		p, err := strconv.Atoi(s)
		if err != nil || p < 1 || p > 65535 {
			invalid = append(invalid, s)
			continue
		}
		out = append(out, p)
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid ports: %s", strings.Join(invalid, ","))
	}
	return out, nil
}

// NormalizeCustom returns a normalized CSV list for custom ports:
// valid ports only, duplicates removed, sorted ascending.
func NormalizeCustom(raw string) (string, error) {
	list, err := parseList(raw)
	if err != nil {
		return "", err
	}

	set := make(map[int]struct{}, len(list))
	for _, p := range list {
		set[p] = struct{}{}
	}

	out := make([]int, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Ints(out)

	parts := make([]string, 0, len(out))
	for _, p := range out {
		parts = append(parts, strconv.Itoa(p))
	}
	return strings.Join(parts, ","), nil
}
