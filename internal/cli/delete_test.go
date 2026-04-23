package cli

import (
	"testing"
	"time"

	"github.com/joncombe/tagbackup/internal/objectkey"
	"github.com/joncombe/tagbackup/internal/store"
)

func mkObj(name string, tsMs int64) store.Object {
	return store.Object{
		Key:    name,
		Size:   1,
		Parsed: &objectkey.Parsed{Timestamp: tsMs, RawTags: "t", Tags: []string{"t"}, DisplayName: name},
	}
}

func TestFilterAge_NewerThan(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	cutoff := now.Add(-2 * 24 * time.Hour).UnixMilli()
	in := []store.Object{
		mkObj("recent", cutoff+1),
		mkObj("boundary", cutoff),
		mkObj("old", cutoff-1),
	}
	got := filterAge(in, now, 2*24*time.Hour, true)
	if len(got) != 1 || got[0].Parsed.DisplayName != "recent" {
		t.Errorf("got %v", got)
	}
}

func TestFilterAge_OlderThan(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	cutoff := now.Add(-2 * 24 * time.Hour).UnixMilli()
	in := []store.Object{
		mkObj("recent", cutoff+1),
		mkObj("boundary", cutoff),
		mkObj("old", cutoff-1),
	}
	got := filterAge(in, now, 2*24*time.Hour, false)
	if len(got) != 1 || got[0].Parsed.DisplayName != "old" {
		t.Errorf("got %v", got)
	}
}

func TestParseTagCSV(t *testing.T) {
	got, err := parseTagCSV("b,a")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "b" || got[1] != "a" {
		t.Errorf("want [b a], got %v", got)
	}
	bad := []string{"", "a,,b", "a b,c", "a-b", "a!"}
	for _, s := range bad {
		if _, err := parseTagCSV(s); err == nil {
			t.Errorf("%q: expected error", s)
		}
	}
}

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		0:               "0 B",
		123:             "123 B",
		1024:            "1.0 KiB",
		1536:            "1.5 KiB",
		1024 * 1024:     "1.0 MiB",
		1024*1024*1024 + 1024*1024*512: "1.5 GiB",
	}
	for in, want := range cases {
		got := humanBytes(in)
		if got != want {
			t.Errorf("humanBytes(%d)=%q, want %q", in, got, want)
		}
	}
}
