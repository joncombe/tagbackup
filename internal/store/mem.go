package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/joncombe/tagbackup/internal/objectkey"
)

// Mem is an in-memory [ObjectStore] for tests. It is not safe for concurrent use.
type Mem struct {
	Bucket string
	Prefix string
	// Objects maps full S3 key -> object data.
	Objects  map[string][]byte
	LogDebug *slog.Logger
}

func (m *Mem) SetDebugLog(l *slog.Logger) { m.LogDebug = l }

// NewMem returns an empty in-memory store with optional prefix.
func NewMem(bucket, prefix string) *Mem {
	return &Mem{
		Bucket:  bucket,
		Prefix:  prefix,
		Objects: make(map[string][]byte),
	}
}

func (m *Mem) relKey(full string) (string, error) {
	np := normalizedPrefix(m.Prefix)
	if np == "" {
		return full, nil
	}
	if !strings.HasPrefix(full, np) {
		return "", fmt.Errorf("key %q does not have expected prefix", full)
	}
	return strings.TrimPrefix(full, np), nil
}

// ListObjectsAll implements [ObjectStore].
func (m *Mem) ListObjectsAll(ctx context.Context, ev func(map[string]struct{}) bool) ([]Object, error) {
	_ = ctx
	var keys []string
	for k := range m.Objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []Object
	for _, key := range keys {
		rel, err := m.relKey(key)
		if err != nil {
			continue
		}
		p, err := objectkey.Parse(rel)
		if err != nil {
			if m.LogDebug != nil {
				m.LogDebug.Debug("skip non-conforming key", "key", key)
			}
			continue
		}
		if ev != nil && !ev(p.TagSet()) {
			continue
		}
		sz := int64(len(m.Objects[key]))
		out = append(out, Object{Key: key, Parsed: p, Size: sz})
	}
	return out, nil
}

// Upload implements [ObjectStore].
func (m *Mem) Upload(ctx context.Context, key string, r io.Reader, _ int64) error {
	_ = ctx
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if m.Objects == nil {
		m.Objects = make(map[string][]byte)
	}
	m.Objects[key] = b
	return nil
}

// GetObjectReader implements [ObjectStore].
func (m *Mem) GetObjectReader(_ context.Context, key string) (io.ReadCloser, int64, error) {
	b, ok := m.Objects[key]
	if !ok {
		return nil, 0, fmt.Errorf("object not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
}

// DeleteObject implements [ObjectStore].
func (m *Mem) DeleteObject(_ context.Context, key string) error {
	if m.Objects == nil {
		return nil
	}
	delete(m.Objects, key)
	return nil
}

// BuildPushKey implements [ObjectStore].
func (m *Mem) BuildPushKey(original string, tagList []string) (string, error) {
	ts := time.Now().UTC().UnixMilli()
	np := normalizedPrefix(m.Prefix)
	rel, err := objectkey.BuildKey("", original, tagList, ts)
	if err != nil {
		return "", err
	}
	if np != "" {
		return np + rel, nil
	}
	return rel, nil
}
