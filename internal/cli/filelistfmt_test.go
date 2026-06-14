package cli

import (
	"strings"
	"testing"

	"github.com/joncombe/tagbackup/internal/objectkey"
	"github.com/joncombe/tagbackup/internal/store"
)

func obj(name, tags string, ts, size int64) store.Object {
	return store.Object{
		Key:  "k",
		Size: size,
		Parsed: &objectkey.Parsed{
			Timestamp:   ts,
			RawTags:     tags,
			Tags:        strings.Split(tags, ","),
			DisplayName: name,
		},
	}
}

func TestFileListFmt_HeaderMinNameWidth(t *testing.T) {
	f := newFileListFmt([]store.Object{obj("x", "maindb", 1_700_000_000_000, 1024)})
	if f.nameWidth != len("FILENAME") {
		t.Fatalf("nameWidth=%d, want %d", f.nameWidth, len("FILENAME"))
	}
	got := f.Header()
	if !strings.HasPrefix(got, "TIMESTAMP") || !strings.HasSuffix(got, "TAGS") {
		t.Fatalf("Header()=%q, unexpected shape", got)
	}
	if !strings.Contains(got, "FILENAME") || !strings.Contains(got, "SIZE") {
		t.Fatalf("Header()=%q, missing column labels", got)
	}
}

func TestFileListFmt_HeaderExpandsForLongNames(t *testing.T) {
	long := "very-long-backup-filename.sql"
	f := newFileListFmt([]store.Object{obj(long, "maindb", 1_700_000_000_000, 1024)})
	if f.nameWidth != len(long) {
		t.Errorf("nameWidth=%d, want %d", f.nameWidth, len(long))
	}
}

func TestFileListFmt_Row(t *testing.T) {
	// 2023-11-14 22:13:20 UTC
	ts := int64(1_700_000_000_000)
	f := newFileListFmt([]store.Object{obj("dump.sql", "maindb,nightly", ts, 1<<20)})
	got := f.Row(obj("dump.sql", "maindb,nightly", ts, 1<<20))
	want := "2023-11-14 22:13:20Z     1.0 MiB  dump.sql  [maindb,nightly]"
	if got != want {
		t.Errorf("Row()=%q, want %q", got, want)
	}
}
