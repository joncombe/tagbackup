package cli

import (
	"fmt"
	"os"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/spf13/cobra"
)

func newRoot() *cobra.Command {
	g := &Runtime{ColorWanted: true}

	root := &cobra.Command{
		Use:   "tagbackup",
		Short: "Manage files on S3-compatible buckets with tags",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if v, _ := cmd.Root().PersistentFlags().GetBool("version"); v {
				g.printVersion()
				os.Exit(0)
			}
			g.Ctx = cmd.Context()
			if os.Getenv("NO_COLOR") != "" {
				g.ColorWanted = false
			}
			if g.NoColor {
				g.ColorWanted = false
			}
			if g.Verbose && g.Quiet {
				return exitUsage("tagbackup", "--verbose and --quiet are mutually exclusive")
			}
			g.InitLogging()
			if path, perr := config.ResolvePath(g.ConfigPath); perr == nil && config.WarnLoosePerms(path) {
				_, _ = fmt.Fprintln(os.Stderr, "tagbackup: warning: config file permissions are not 600; consider: chmod 600", path)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	g.bindGlobalFlags(root)
	root.PersistentFlags().Bool("version", false, "print version and exit")

	root.AddCommand(
		g.cmdPush(),
		g.cmdPull(),
		g.cmdFiles(),
		g.cmdDelete(),
		g.cmdTags(),
		g.cmdBucket(),
	)

	return root
}
