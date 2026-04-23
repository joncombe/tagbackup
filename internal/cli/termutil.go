package cli

import (
	"os"

	"golang.org/x/term"
)

// StderrIsTTY returns whether stderr is a terminal.
func StderrIsTTY() bool { return term.IsTerminal(int(os.Stderr.Fd())) }

// StdinIsTTY whether stdin is a terminal.
func StdinIsTTY() bool { return term.IsTerminal(int(os.Stdin.Fd())) }

// useColor returns true when colour output is wanted and stderr is a TTY.
func (g *Runtime) useColor() bool { return g.ColorWanted && StderrIsTTY() }

// ansi wraps s with a colour escape when the runtime wants colour.
func (g *Runtime) ansi(code, s string) string {
	if !g.useColor() {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (g *Runtime) red(s string) string    { return g.ansi("31", s) }
func (g *Runtime) green(s string) string  { return g.ansi("32", s) }
func (g *Runtime) yellow(s string) string { return g.ansi("33", s) }

// showProgressBar reports whether a progress bar should be shown on stderr.
func (g *Runtime) showProgressBar() bool { return !g.Quiet && StderrIsTTY() }
