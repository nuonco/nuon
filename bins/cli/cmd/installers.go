package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/installers"
	"github.com/spf13/cobra"
)

func (c *cli) installersCmd() *cobra.Command {
	installsCmds := &cobra.Command{
		Use:               "installers",
		Short:             "Manage installers",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installers",
		Long:    "List all installers",
		Run: func(cmd *cobra.Command, _ []string) {
			svc := installers.New(c.apiClient, c.cfg)
			svc.List(cmd.Context(), PrintJSON)
		},
	}
	installsCmds.AddCommand(listCmd)

	return installsCmds
}
