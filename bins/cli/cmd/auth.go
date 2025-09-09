package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/auth"
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
		Use:   "login",
		Short: "Login to Nuon",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := auth.New(c.apiClient)
			return svc.Login(cmd.Context(), c.cfg)
		}),
	}

	// Add logout subcommand
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Nuon",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.logout(cmd.Context())
		}),
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)

	return authCmd
}

// logout handles the logout functionality by clearing the API token and URL
func (c *cli) logout(ctx context.Context) error {
	// Clear the API token and URL from config
	c.cfg.Set("api_token", "")
	c.cfg.Set("api_url", "")

	// Write the updated config to file
	if err := c.cfg.WriteConfig(); err != nil {
		return err
	}

	// Print success message
	cmd := &cobra.Command{}
	cmd.Printf("✅ Successfully logged out.\n")

	return nil
}

