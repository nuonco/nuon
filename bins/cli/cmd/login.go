package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/auth"
	"github.com/spf13/cobra"
)

func (c *cli) loginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:               "login",
		Short:             "Login to Nuon",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, args []string) {
			svc := auth.New(c.apiClient)
			svc.Login(cmd.Context(), c.cfg)
		},
	}

	return loginCmd
}
