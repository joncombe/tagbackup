package tagexpr

import (
	"testing"
)

func set(tags ...string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, t := range tags {
		m[t] = struct{}{}
	}
	return m
}

func TestParse_basic(t *testing.T) {
	ev, err := Parse("a")
	if err != nil {
		t.Fatal(err)
	}
	if !ev(set("a", "b")) {
		t.Fatal("expected match")
	}
	if ev(set("b")) {
		t.Fatal("expected no match")
	}
}

func TestParse_or(t *testing.T) {
	ev, err := Parse("a|b")
	if err != nil {
		t.Fatal(err)
	}
	if !ev(set("a")) || !ev(set("b")) {
		t.Fatal("or")
	}
	if ev(set("c")) {
		t.Fatal("or c")
	}
}

func TestParse_and(t *testing.T) {
	ev, err := Parse("a+b")
	if err != nil {
		t.Fatal(err)
	}
	if !ev(set("a", "b")) {
		t.Fatal("and ab")
	}
	if ev(set("a")) {
		t.Fatal("and a only")
	}
}

func TestParse_not(t *testing.T) {
	ev, err := Parse("-a")
	if err != nil {
		t.Fatal(err)
	}
	if ev(set("a")) {
		t.Fatal("not a")
	}
	if !ev(set("b")) {
		t.Fatal("not b")
	}
}

func TestParse_complex(t *testing.T) {
	ev, err := Parse("(-a+-b)|c")
	if err != nil {
		t.Fatal(err)
	}
	if !ev(set("c")) {
		t.Fatal("c ok")
	}
	if !ev(set("x")) { // has neither a nor b, so first branch... (-a) is true, (-b) is true, a+b? wait: -a + -b means (not a) AND (not b)
		t.Fatalf("x should match (no a no b) -> (-a) true, (-b) true")
	}
	if !ev(set("a", "b", "c")) { // c exists -> or matches
		t.Fatal("abc c")
	}
}

func TestParse_whitespace(t *testing.T) {
	_, err := Parse("a b")
	if err == nil {
		t.Fatal("expected error")
	}
}
