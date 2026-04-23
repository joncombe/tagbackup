package store

import (
	"context"
	"testing"

	"github.com/joncombe/tagbackup/internal/config"
)

func TestNewObjectStoreFromAlias_unknown(t *testing.T) {
	cfg := &config.Cfg{Version: 1, Buckets: map[string]config.Bucket{}}
	_, err := NewObjectStoreFromAlias(context.Background(), cfg, "missing", nil)
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}
}
