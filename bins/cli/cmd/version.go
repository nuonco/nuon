package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/version"
)

func (c *cli) versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:               "version",
		Short:             "Show the version of the CLI you are using",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := version.New()
			return svc.Version(cmd.Context(), PrintJSON)
		}),
		GroupID: HelpGroup.ID,
	}

	return versionCmd
}
