package tagexpr

import "testing"

func TestParse_EmptyAndInvalid(t *testing.T) {
	for _, s := range []string{
		"",
		"+",
		"|",
		"a+",
		"a|",
		"(a",
		"a)",
		"a-b",
		"()",
		"(a+)",
		"--a",
	} {
		if _, err := Parse(s); err == nil {
			t.Errorf("%q: expected parse error", s)
		}
	}
}

func TestParse_DoubleNegation(t *testing.T) {
	ev, err := Parse("-(-a)")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !ev(set("a")) {
		t.Error("double-negated a should match {a}")
	}
	if ev(set("b")) {
		t.Error("double-negated a should not match {b}")
	}
}

func TestParse_PrecedenceAndBeforeOr(t *testing.T) {
	ev, err := Parse("a+b|c+d")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !ev(set("a", "b")) {
		t.Error("a+b should match")
	}
	if !ev(set("c", "d")) {
		t.Error("c+d should match")
	}
	if ev(set("a", "c")) {
		t.Error("a+c should not match (a without b on lhs, c without d on rhs)")
	}
}

func TestParse_OnlyHyphens(t *testing.T) {
	if _, err := Parse("-"); err == nil {
		t.Error("expected error for bare -")
	}
}

func TestParse_Grouping(t *testing.T) {
	ev, err := Parse("(a|b)+c")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !ev(set("a", "c")) || !ev(set("b", "c")) {
		t.Error("(a|b)+c should match {a,c} and {b,c}")
	}
	if ev(set("a", "b")) {
		t.Error("(a|b)+c should not match without c")
	}
}

func TestParse_Invalid_TagChars(t *testing.T) {
	for _, s := range []string{"a!", "a.b", "a@b"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("%q: expected error", s)
		}
	}
}
