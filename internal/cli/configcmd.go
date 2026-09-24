package cli

import (
	"fmt"
	"os"

	"github.com/joncombe/tagbackup/internal/config"
	"github.com/spf13/cobra"
)

func (g *Runtime) cmdConfig() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Inspect the config file",
	}
	c.AddCommand(g.cmdConfigPath())
	return c
}

func (g *Runtime) cmdConfigPath() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ResolvePath(g.ConfigPath)
			if err != nil {
				return exitConfig("config path", err)
			}
			_, _ = fmt.Fprintln(os.Stdout, path)
			return nil
		},
	}
}
