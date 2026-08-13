package cmd

import (
	"github.com/spf13/cobra"
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
			svc := c.roles
			return svc.ListRoles(cmd.Context(), PrintJSON)
		}),
	}
	rolesCmd.AddCommand(listCmd)

	return rolesCmd
}
