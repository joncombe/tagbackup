package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/joncombe/tagbackup/internal/server"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdServe() *cobra.Command {
	var port int
	var noOpen bool
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve a local web UI for browsing, uploading, downloading, and deleting files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.runServe(port, noOpen)
		},
	}
	cmd.Flags().IntVar(&port, "port", 3000, "port to listen on (localhost only)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "do not open the browser automatically")
	return cmd
}

func (g *Runtime) runServe(port int, noOpen bool) error {
	const name = "serve"
	if port < 1 || port > 65535 {
		return exitUsage(name, "--port must be between 1 and 65535")
	}

	srv := server.New(server.Options{
		ConfigPath: g.ConfigPath,
		Debug:      g.debugLogger(),
		Version:    version,
	})

	// Bind to localhost only: this exposes bucket filenames and must never be
	// reachable from the network.
	ln, err := srv.Listen("127.0.0.1", port)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			return exitErrMsg(name, "port %d is already in use; choose another with --port", port)
		}
		return exitErr(name, err)
	}

	url := fmt.Sprintf("http://%s", ln.Addr().String())
	_, _ = fmt.Fprintf(os.Stderr, "tagbackup serve: listening on %s (press Ctrl+C to stop)\n", url)

	if !noOpen {
		if oerr := openBrowser(url); oerr != nil && g.Verbose {
			g.Log.Debug("could not open browser", "err", oerr)
		}
	}

	if serr := srv.Serve(g.Ctx, ln); serr != nil {
		return exitErr(name, serr)
	}
	return nil
}

// openBrowser opens url in the user's default browser, best-effort.
func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	return exec.Command(cmd, args...).Start()
}
