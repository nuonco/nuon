package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/installers"
)

func (c *cli) installersCmd() *cobra.Command {
	installsCmds := &cobra.Command{
		Use:               "installers",
		Short:             "Manage installers",
		Aliases:           []string{"i"},
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installers",
		Long:    "List all installers",
		Run: c.run(func(cmd *cobra.Command, _ []string) error {
			svc := installers.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), PrintJSON)
		}),
	}
	installsCmds.AddCommand(listCmd)

	return installsCmds
}
