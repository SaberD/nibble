package ports

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type portRange struct {
	start int
	end   int
}

// normalizeRanges sorts and merges overlapping ranges
func normalizeRanges(ranges []portRange) string {
	if len(ranges) == 0 {
		return ""
	}

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start != ranges[j].start {
			return ranges[i].start < ranges[j].start
		}
		return ranges[i].end < ranges[j].end
	})

	merged := make([]portRange, 0, len(ranges))
	for _, curr := range ranges {
		if len(merged) == 0 {
			merged = append(merged, curr)
			continue
		}
		lastIdx := len(merged) - 1
		last := merged[lastIdx]
		if curr.start <= last.end {
			if curr.end > last.end {
				merged[lastIdx].end = curr.end
			}
			continue
		}
		merged = append(merged, curr)
	}

	out := make([]string, 0, len(merged))
	for _, m := range merged {
		out = append(out, formatToken(m.start, m.end))
	}
	return strings.Join(out, ",")
}

// formatToken renders a single port or an inclusive range
func formatToken(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}

// parseTokenBounds parses "port" or "start-end" and returns inclusive bounds
func parseTokenBounds(raw string) (int, int, error) {
	if strings.Count(raw, "-") == 0 {
		p, err := strconv.Atoi(raw)
		if err != nil || p < 1 || p > 65535 {
			return 0, 0, fmt.Errorf("invalid port")
		}
		return p, p, nil
	}
	if strings.Count(raw, "-") != 1 {
		return 0, 0, fmt.Errorf("invalid range")
	}

	parts := strings.SplitN(raw, "-", 2)
	startRaw := strings.TrimSpace(parts[0])
	endRaw := strings.TrimSpace(parts[1])
	if startRaw == "" || endRaw == "" {
		return 0, 0, fmt.Errorf("invalid range")
	}

	start, err := strconv.Atoi(startRaw)
	if err != nil || start < 1 || start > 65535 {
		return 0, 0, fmt.Errorf("invalid range")
	}
	end, err := strconv.Atoi(endRaw)
	if err != nil || end < 1 || end > 65535 {
		return 0, 0, fmt.Errorf("invalid range")
	}
	if start > end {
		return 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, nil
}
