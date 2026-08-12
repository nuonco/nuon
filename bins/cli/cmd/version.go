package cmd

import (
	"github.com/spf13/cobra"
)

func (c *cli) versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:               "version",
		Short:             "Show the version of the CLI you are using",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.version
			return svc.Version(cmd.Context(), PrintJSON)
		}),
		GroupID: HelpGroup.ID,
	}

	return versionCmd
}
