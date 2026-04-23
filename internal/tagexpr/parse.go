package tagexpr

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/joncombe/tagbackup/internal/objectkey"
)

type parser struct {
	s   string
	i   int
	err error
}

// Parse returns an evaluator for a tag expression, or a syntax error.
// Whitespace in the input is a syntax error.
func Parse(s string) (func(map[string]struct{}) bool, error) {
	if strings.ContainsFunc(s, unicode.IsSpace) {
		return nil, fmt.Errorf("whitespace is not allowed in tag expression")
	}
	p := &parser{s: s}
	n := p.parseOr()
	if p.err != nil {
		return nil, p.err
	}
	if p.i < len(p.s) {
		return nil, fmt.Errorf("unexpected %q at position %d", p.s[p.i], p.i)
	}
	if n == nil {
		return nil, fmt.Errorf("empty expression")
	}
	return n.eval, nil
}

func (p *parser) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

func (p *parser) parseOr() node {
	if p.err != nil {
		return nil
	}
	left := p.parseAnd()
	if p.err != nil {
		return nil
	}
	for p.peek() == '|' {
		p.next()
		right := p.parseAnd()
		if p.err != nil {
			return nil
		}
		left = &orNode{left, right}
	}
	return left
}

func (p *parser) parseAnd() node {
	if p.err != nil {
		return nil
	}
	left := p.parseUnary()
	if p.err != nil {
		return nil
	}
	for p.peek() == '+' {
		p.next()
		right := p.parseUnary()
		if p.err != nil {
			return nil
		}
		left = &andNode{left, right}
	}
	return left
}

func (p *parser) parseUnary() node {
	if p.err != nil {
		return nil
	}
	if p.peek() == '-' {
		p.next()
		inner := p.parsePrimary()
		if p.err != nil {
			return nil
		}
		return &notNode{inner}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() node {
	if p.err != nil {
		return nil
	}
	switch p.peek() {
	case 0:
		p.setErr(fmt.Errorf("unexpected end of expression"))
		return nil
	case '(':
		p.next()
		inner := p.parseOr()
		if p.err != nil {
			return nil
		}
		if p.peek() != ')' {
			p.setErr(fmt.Errorf("expected ')' at position %d", p.i))
			return nil
		}
		p.next()
		return inner
	default:
		return p.parseTag()
	}
}

func (p *parser) parseTag() node {
	start := p.i
	for p.i < len(p.s) && objectkey.IsTagChar(p.s[p.i]) {
		p.i++
	}
	if p.i == start {
		if p.i < len(p.s) {
			p.setErr(fmt.Errorf("unexpected %q in tag at position %d", p.s[p.i], p.i))
		} else {
			p.setErr(fmt.Errorf("expected tag at position %d", p.i))
		}
		return nil
	}
	return &tagNode{p.s[start:p.i]}
}

func (p *parser) peek() byte {
	if p.i >= len(p.s) {
		return 0
	}
	return p.s[p.i]
}

func (p *parser) next() {
	if p.i < len(p.s) {
		p.i++
	}
}
