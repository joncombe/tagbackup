package store

import (
	"context"
	"strings"
	"testing"

	"github.com/joncombe/tagbackup/internal/tagexpr"
)

func TestMem_ListObjectsAll_tagFilter(t *testing.T) {
	m := NewMem("b", "p/")
	_ = m.Upload(context.Background(), "p/1000000000000-a,b-f.txt", strings.NewReader("a"), 1)
	_ = m.Upload(context.Background(), "p/2000000000000-b,c-g.bin", strings.NewReader("b"), 1)

	ev, err := tagexpr.Parse("a+b")
	if err != nil {
		t.Fatal(err)
	}
	objs, err := m.ListObjectsAll(context.Background(), ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 1 {
		t.Fatalf("want 1 match, got %d", len(objs))
	}
	if objs[0].Parsed.DisplayName != "f.txt" {
		t.Fatalf("wrong object: %#v", objs[0])
	}
}
