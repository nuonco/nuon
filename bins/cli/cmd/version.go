package cmd

import (
	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerVersion() cobra.Command {
	versionCmd := &cobra.Command{
		Use: "version",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			ui.Line(ctx, "%s\n", "development")
			return nil
		},
	}

	return *versionCmd
}
