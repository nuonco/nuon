package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/auth"
)

func (c *cli) loginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Deprecated:        "Use `nuon auth login` instead",
		Use:               "login",
		Short:             "Login to Nuon (deprecated)",
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := auth.New(c.apiClient)
			return svc.Login(cmd.Context(), c.cfg)
		}),
		GroupID: AdditionalGroup.ID,
	}

	return loginCmd
}
