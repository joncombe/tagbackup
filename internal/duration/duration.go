package duration

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseRelativeUTC parses values like "2d", "30m" and returns a duration.
// Units: s, m, h, d, w (seconds, minutes, hours, days, weeks).
func ParseRelativeUTC(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	u := s[len(s)-1]
	num := s[:len(s)-1]
	if num == "" {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	if n < 0 {
		return 0, fmt.Errorf("duration must be non-negative")
	}
	switch u {
	case 's':
		return time.Duration(n) * time.Second, nil
	case 'm':
		return time.Duration(n) * time.Minute, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit in %q (use s, m, h, d, w)", s)
	}
}
