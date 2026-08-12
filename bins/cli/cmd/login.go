package cmd

import (
	"github.com/spf13/cobra"
)

func (c *cli) loginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Deprecated:        "Use `nuon auth login` instead",
		Use:               "login",
		Short:             "Login to Nuon (deprecated)",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			svc := c.auth
			return svc.Login(cmd.Context())
		}),
		GroupID: AdditionalGroup.ID,
	}

	return loginCmd
}
