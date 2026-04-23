package store

import (
	"context"
	"io"
	"log/slog"
)

// ObjectStore abstracts S3 operations used by the CLI, so the core logic can
// be exercised with a fake in tests (see Mem).
type ObjectStore interface {
	ListObjectsAll(ctx context.Context, ev func(map[string]struct{}) bool) ([]Object, error)
	Upload(ctx context.Context, key string, r io.Reader, size int64) error
	GetObjectReader(ctx context.Context, key string) (io.ReadCloser, int64, error)
	DeleteObject(ctx context.Context, key string) error
	BuildPushKey(original string, tagList []string) (string, error)
	// SetDebugLog enables DEBUG logs for ListObjectsAll (skipped keys); no-op is allowed.
	SetDebugLog(l *slog.Logger)
}

var _ ObjectStore = (*S3)(nil)
var _ ObjectStore = (*Mem)(nil)
