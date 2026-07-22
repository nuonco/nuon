package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/roles"
)

func (c *cli) rolesCmd() *cobra.Command {
	rolesCmd := &cobra.Command{
		Use:               "roles",
		Short:             "View roles assignable to members and service accounts",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List roles assignable to members and service accounts",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := roles.New(c.apiClient, c.cfg)
			return svc.ListRoles(cmd.Context(), PrintJSON)
		}),
	}
	rolesCmd.AddCommand(listCmd)

	return rolesCmd
}
