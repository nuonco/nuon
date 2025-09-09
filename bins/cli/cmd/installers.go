package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/installers"
)

func (c *cli) installersCmd() *cobra.Command {
	var (
		offset int
		limit  int
	)

	installsCmds := &cobra.Command{
		Use:               "installers",
		Short:             "Manage app installers",
		Aliases:           []string{"i"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installers",
		Long:    "List all installers",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installers.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum installers to return")
	installsCmds.AddCommand(listCmd)

	return installsCmds
}
