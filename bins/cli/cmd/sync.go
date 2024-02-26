package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/spf13/cobra"
)

func (c *cli) syncCmd() *cobra.Command {
	var (
		format string
		all    bool
		file   string
	)
	syncCmd := &cobra.Command{
		Use:               "sync",
		Short:             "Sync all .nuon.toml config files in the current directory.",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := apps.New(c.apiClient, c.cfg)
			svc.Sync(cmd.Context(), all, file, format, PrintJSON)
		},
	}

	syncCmd.Flags().StringVarP(&format, "format", "", "toml", "Config file format (toml, json are allowed)")
	syncCmd.Flags().StringVarP(&file, "file", "", "", "Config file to sync")
	syncCmd.Flags().BoolVarP(&all, "all", "", true, "sync all config files found")

	return syncCmd
}
