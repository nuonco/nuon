package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/auth"
)

func (c *cli) loginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Deprecated:        "Use `nuon auth login` instead",
		Use:               "login",
		Short:             "Login to Nuon (deprecated)",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := auth.New(c.apiClient, c.cfg)
			return svc.Login(cmd.Context())
		}),
		GroupID: AdditionalGroup.ID,
	}

	return loginCmd
}
