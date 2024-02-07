package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/spf13/cobra"
)

func (c *cli) versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:               "version",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, _ []string) {
			svc := version.New()
			svc.Version(cmd.Context(), PrintJSON)
		},
	}

	return versionCmd
}
