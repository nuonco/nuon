package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/apps"
	"github.com/powertoolsdev/mono/bins/cli/internal/installs"
	"github.com/powertoolsdev/mono/bins/cli/internal/orgs"
)

func (c *cli) configCmd() *cobra.Command {
	var (
		id        string
		appID     string
		installID string
	)

	configCmd := &cobra.Command{
		// TODO(ja): fix config file bugs before re-enabling this
		Hidden:            true,
		Use:               "config",
		Short:             "Configure the CLI",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           CoreGroup.ID,
	}

	// Add org subcommand
	orgCmd := &cobra.Command{
		Use:   "org",
		Short: "Select your current org",
		Long:  "Select your current org from a list or by org ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), id, PrintJSON)
		}),
	}
	orgCmd.Flags().StringVar(&id, "org", "", "The ID of the org you want to use")
	configCmd.AddCommand(orgCmd)

	// Add app subcommand
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Select your current app",
		Long:  "Select your current app from a list or by app ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := apps.New(c.v, c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), appID, PrintJSON)
		}),
	}
	appCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	configCmd.AddCommand(appCmd)

	// Add install subcommand
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Select your current install",
		Long:  "Select your current install from a list or by install ID",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := installs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), appID, installID, PrintJSON)
		}),
	}
	installCmd.Flags().StringVar(&installID, "install", "", "The ID of the install you want to use")
	installCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	configCmd.AddCommand(installCmd)

	// Add clear subcommand
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear configuration except token",
		Long:  "Remove all configuration settings except for the API token",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.clearConfig(cmd.Context())
		}),
	}
	configCmd.AddCommand(clearCmd)

	return configCmd
}

// clearConfig clears all configuration settings except the API token
func (c *cli) clearConfig(ctx context.Context) error {
	// Get current API token to preserve it
	apiToken := c.cfg.GetString("api_token")

	// Clear the configuration
	c.cfg.Set("org_id", "")
	c.cfg.Set("app_id", "")
	c.cfg.Set("install_id", "")

	// Restore the API token
	c.cfg.Set("api_token", apiToken)

	// Write the updated config to file
	if err := c.cfg.WriteConfig(); err != nil {
		return err
	}

	// Print success message
	cmd := &cobra.Command{}
	cmd.Printf("✅ Configuration cleared.\n")

	return nil
}
