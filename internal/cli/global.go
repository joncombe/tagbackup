package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/spf13/cobra"
)

// version is set by -ldflags.
var version = "dev"

// Runtime holds per-invocation state (flags + loaded config + logger).
type Runtime struct {
	Ctx  context.Context
	Log  *slog.Logger
	Cfg  *config.Cfg
	Conf string // resolved config path

	ConfigPath   string
	Verbose      bool
	Quiet        bool
	NonInter     bool
	NoColor      bool
	ColorWanted  bool // after NO_COLOR and --no-color
}

// InitLogging sets up slog to stderr; level depends on -v and -q.
func (g *Runtime) InitLogging() {
	lvl := slog.LevelInfo
	if g.Verbose {
		lvl = slog.LevelDebug
	}
	if g.Quiet {
		lvl = slog.LevelWarn
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	g.Log = slog.New(h)
}

// debugLogger returns the runtime logger only when --verbose is active.
// Used by the store layer to switch on SDK retry + credential-path logs.
func (g *Runtime) debugLogger() *slog.Logger {
	if g.Verbose {
		return g.Log
	}
	return nil
}

// bindGlobalFlags registers flags on the root command.
func (g *Runtime) bindGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&g.ConfigPath, "config", "", "path to config file (default: per-user path)")
	cmd.PersistentFlags().BoolVarP(&g.Verbose, "verbose", "v", false, "log debug to stderr")
	cmd.PersistentFlags().BoolVarP(&g.Quiet, "quiet", "q", false, "suppress most log output")
	cmd.PersistentFlags().BoolVar(&g.NonInter, "non-interactive", false, "fail instead of prompting")
	cmd.PersistentFlags().BoolVar(&g.NoColor, "no-color", false, "disable colored output")
}
