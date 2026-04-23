package tagexpr

// node is a tag expression AST node.
type node interface {
	eval(have map[string]struct{}) bool
}

type orNode struct{ a, b node }
type andNode struct{ a, b node }
type notNode struct{ inner node }
type tagNode struct{ name string }

func (n *orNode) eval(have map[string]struct{}) bool  { return n.a.eval(have) || n.b.eval(have) }
func (n *andNode) eval(have map[string]struct{}) bool { return n.a.eval(have) && n.b.eval(have) }
func (n *notNode) eval(have map[string]struct{}) bool {
	return !n.inner.eval(have)
}
func (n *tagNode) eval(have map[string]struct{}) bool { _, ok := have[n.name]; return ok }
