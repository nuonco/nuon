package cmd

import (
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/services/auth"
)

func (c *cli) authCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:               "auth",
		Short:             "Login and logout commands",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           CoreGroup.ID,
	}

	// Add login subcommand
	loginCmd := &cobra.Command{
		Use:         "login",
		Short:       "Login to Nuon",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := auth.New(c.apiClient, c.cfg)
			return svc.Login(cmd.Context())
		}),
	}

	// Add logout subcommand
	logoutCmd := &cobra.Command{
		Use:         "logout",
		Short:       "Logout from Nuon",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := auth.New(c.apiClient, c.cfg)
			return svc.Logout(cmd.Context())
		}),
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)

	return authCmd
}
