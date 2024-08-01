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
		Run: c.run(func(cmd *cobra.Command, args []string) error {
			svc := auth.New(c.apiClient)
			return svc.Login(cmd.Context(), c.cfg)
		}),
	}

	return loginCmd
}
