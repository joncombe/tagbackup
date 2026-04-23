package objectkey

import (
	"strings"
	"testing"
)

func TestParse_Basic(t *testing.T) {
	p, err := Parse("1776831788343-mytag1,mytag2-my-file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Timestamp != 1776831788343 {
		t.Errorf("timestamp: got %d", p.Timestamp)
	}
	if p.DisplayName != "my-file.txt" {
		t.Errorf("name: got %q", p.DisplayName)
	}
	if p.RawTags != "mytag1,mytag2" {
		t.Errorf("raw tags: got %q", p.RawTags)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "mytag1" || p.Tags[1] != "mytag2" {
		t.Errorf("tags: got %v", p.Tags)
	}
}

func TestParse_FilenameWithHyphens(t *testing.T) {
	p, err := Parse("1776831788343-a-my-very-long-file-name.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DisplayName != "my-very-long-file-name.txt" {
		t.Errorf("name: got %q", p.DisplayName)
	}
}

func TestParse_Invalid(t *testing.T) {
	cases := []string{
		"",
		"nope",
		"abc-tag-name.txt",                  // non-numeric ts
		"123-tag-name.txt",                  // short ts
		"1776831788343--name.txt",           // empty tag segment
		"1776831788343-bad!tag-name.txt",    // invalid tag char
		"1776831788343-tag1,,tag2-name.txt", // empty tag inside csv
	}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("%q: expected error, got nil", c)
		}
	}
}

func TestBuildKey_SortsTags(t *testing.T) {
	got, err := BuildKey("", "hello.txt", []string{"b", "a", "c"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	const want = "1-a,b,c-hello.txt"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestBuildKey_PrefixNormalised(t *testing.T) {
	got, err := BuildKey("some/prefix", "hello.txt", []string{"x"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != "some/prefix/1-x-hello.txt" {
		t.Errorf("got %q", got)
	}
	got2, _ := BuildKey("some/prefix/", "hello.txt", []string{"x"}, 1)
	if got != got2 {
		t.Errorf("prefix trailing slash not idempotent: %q vs %q", got, got2)
	}
}

func TestBuildKey_Limits(t *testing.T) {
	big := strings.Repeat("a", MaxOriginalNameBytes+1)
	if _, err := BuildKey("", big, []string{"x"}, 1); err == nil {
		t.Error("expected error for oversize original filename")
	}
	// Oversize overall key via very long prefix.
	bigPrefix := strings.Repeat("p", MaxKeyBytes)
	if _, err := BuildKey(bigPrefix, "x.txt", []string{"t"}, 1); err == nil {
		t.Error("expected error for oversize assembled key")
	}
}

func TestBuildKey_NoTags(t *testing.T) {
	if _, err := BuildKey("", "x.txt", nil, 1); err == nil {
		t.Error("expected error for zero tags")
	}
	if _, err := BuildKey("", "x.txt", []string{""}, 1); err == nil {
		t.Error("expected error for empty tag")
	}
	if _, err := BuildKey("", "x.txt", []string{"bad!tag"}, 1); err == nil {
		t.Error("expected error for invalid tag char")
	}
}

func TestValidateTag(t *testing.T) {
	if err := ValidateTag("abc123"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := ValidateTag(""); err == nil {
		t.Error("empty should fail")
	}
	if err := ValidateTag("a-b"); err == nil {
		t.Error("hyphen should fail")
	}
	if err := ValidateTag("a b"); err == nil {
		t.Error("space should fail")
	}
}
