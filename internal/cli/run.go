package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/joncombe/tagbackup/internal/exitc"
)

// Main runs the CLI and returns a process exit code.
func Main() int {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	var onCancelExit int32 = int32(exitc.SIGINT)
	go func() {
		s := <-sigCh
		if s == syscall.SIGTERM {
			atomic.StoreInt32(&onCancelExit, int32(exitc.SIGTERM))
		} else {
			atomic.StoreInt32(&onCancelExit, int32(exitc.SIGINT))
		}
		cancel()
	}()
	defer signal.Stop(sigCh)

	cmd := newRoot()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		var e *Exit
		if errors.As(err, &e) {
			if e.Human != "" {
				_, _ = fmt.Fprintln(os.Stderr, e.Human)
			}
			return e.Code
		}
		if errors.Is(err, context.Canceled) {
			return int(atomic.LoadInt32(&onCancelExit))
		}
		code, msg := classifyCobraError(err)
		_, _ = fmt.Fprintln(os.Stderr, "tagbackup: "+msg)
		return code
	}
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
		return int(atomic.LoadInt32(&onCancelExit))
	}
	return exitc.OK
}

// classifyCobraError maps Cobra / pflag error strings onto tagbackup exit codes.
// Flag / argument validation problems are usage errors (code 2); anything else
// is a generic error (code 1).
func classifyCobraError(err error) (int, string) {
	msg := err.Error()
	usagePatterns := []string{
		"required flag",
		"unknown flag",
		"unknown shorthand flag",
		"unknown command",
		"invalid argument",
		"flag needs an argument",
		"accepts ",
		"requires ",
	}
	for _, p := range usagePatterns {
		if strings.Contains(msg, p) {
			return exitc.Usage, msg
		}
	}
	return exitc.Err, msg
}
