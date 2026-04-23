package store

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// Exercise the Mem store through a push/list/pull/delete round trip using the
// ObjectStore surface. This is a stand-in for a real end-to-end test; the CLI
// layers just call these same methods against either Mem (in tests) or S3.
func TestMem_RoundTrip(t *testing.T) {
	ctx := context.Background()
	m := NewMem("any-bucket", "nightly/")

	k1, err := m.BuildPushKey("hello.txt", []string{"b", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(k1, "nightly/") {
		t.Errorf("expected prefix, got %q", k1)
	}
	if !strings.Contains(k1, "-a,b-hello.txt") {
		t.Errorf("tags should be sorted in key, got %q", k1)
	}
	if err := m.Upload(ctx, k1, strings.NewReader("body1"), -1); err != nil {
		t.Fatal(err)
	}

	k2, _ := m.BuildPushKey("other.txt", []string{"a"})
	_ = m.Upload(ctx, k2, strings.NewReader("body2"), -1)

	ev, err := ParseTagExpr("a")
	if err != nil {
		t.Fatal(err)
	}
	objs, err := m.ListObjectsAll(ctx, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 2 {
		t.Fatalf("want 2 objects matching tag 'a', got %d", len(objs))
	}

	ev2, _ := ParseTagExpr("b")
	objs2, _ := m.ListObjectsAll(ctx, ev2)
	if len(objs2) != 1 {
		t.Fatalf("want 1 object matching tag 'b', got %d", len(objs2))
	}

	rc, n, err := m.GetObjectReader(ctx, k1)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if n != int64(len("body1")) {
		t.Errorf("size mismatch: got %d", n)
	}
	b, _ := io.ReadAll(rc)
	if !bytes.Equal(b, []byte("body1")) {
		t.Errorf("body mismatch: %q", b)
	}

	if err := m.DeleteObject(ctx, k1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.GetObjectReader(ctx, k1); err == nil {
		t.Error("expected error after delete")
	}
}

func TestMem_SkipsNonConformingKeys(t *testing.T) {
	ctx := context.Background()
	m := NewMem("any-bucket", "")
	// A legit tagbackup object.
	good, _ := m.BuildPushKey("a.txt", []string{"t"})
	_ = m.Upload(ctx, good, strings.NewReader("x"), -1)
	// A foreign object with a non-tagbackup name (e.g. produced by another tool).
	_ = m.Upload(ctx, "some/legacy-upload.bin", strings.NewReader("y"), -1)

	objs, err := m.ListObjectsAll(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 1 || objs[0].Key != good {
		t.Fatalf("legacy object should be skipped; got %v", objs)
	}
}
