package duration

import (
	"testing"
	"time"
)

func TestParseRelativeUTC(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1s", time.Second},
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"2h", 2 * time.Hour},
		{"2d", 48 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"0d", 0},
	}
	for _, c := range cases {
		got, err := ParseRelativeUTC(c.in)
		if err != nil {
			t.Errorf("%s: err %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("%s: got %v want %v", c.in, got, c.want)
		}
	}
}

func TestParseRelativeUTC_Errors(t *testing.T) {
	bad := []string{"", "d", "-1d", "2x", "foo", "1.5d", "w"}
	for _, s := range bad {
		if _, err := ParseRelativeUTC(s); err == nil {
			t.Errorf("%q: expected error", s)
		}
	}
}
