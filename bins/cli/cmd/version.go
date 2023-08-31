package cmd

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
	"github.com/spf13/cobra"
)

func (c *cli) registerVersion(ctx context.Context) cobra.Command {
	versionCmd := &cobra.Command{
		Use: "version",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			ui.Line(ctx, "%s\n", "development")
			return nil
		},
	}

	return *versionCmd
}
