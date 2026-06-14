package cli

import (
	"fmt"
	"time"

	"github.com/joncombe/tagbackup/internal/store"
)

const fileListTimestampFmt = "2006-01-02 15:04:05Z"

type fileListFmt struct {
	nameWidth int
}

func newFileListFmt(objs []store.Object) fileListFmt {
	nameWidth := len("FILENAME")
	for _, o := range objs {
		if n := len(o.Parsed.DisplayName); n > nameWidth {
			nameWidth = n
		}
	}
	return fileListFmt{nameWidth: nameWidth}
}

func (f fileListFmt) Header() string {
	return fmt.Sprintf("%-20s  %10s  %-*s  %s", "TIMESTAMP", "SIZE", f.nameWidth, "FILENAME", "TAGS")
}

func (f fileListFmt) Row(o store.Object) string {
	ts := time.UnixMilli(o.Parsed.Timestamp).UTC().Format(fileListTimestampFmt)
	return fmt.Sprintf("%-20s  %10s  %-*s  [%s]", ts, humanBytes(o.Size), f.nameWidth, o.Parsed.DisplayName, o.Parsed.RawTags)
}
