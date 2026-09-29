package cmd

import (
	"context"

	"github.com/spf13/cobra"
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

	orgCmd := &cobra.Command{
		Use:         "org",
		Short:       "Select your current org",
		Long:        "Select your current org from a list or by org ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.orgs
			return svc.Select(cmd.Context(), id, 0, 50, PrintJSON)
		}),
	}
	orgCmd.Flags().StringVar(&id, "org", "", "The ID of the org you want to use")
	configCmd.AddCommand(orgCmd)

	appCmd := &cobra.Command{
		Use:         "app",
		Short:       "Select your current app",
		Long:        "Select your current app from a list or by app ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.apps
			return svc.Select(cmd.Context(), appID, PrintJSON)
		}),
	}
	appCmd.Flags().StringVar(&appID, "app", "", "The ID of the app you want to use")
	configCmd.AddCommand(appCmd)

	installCmd := &cobra.Command{
		Use:         "install",
		Short:       "Select your current install",
		Long:        "Select your current install from a list or by install ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.installs
			return svc.Select(cmd.Context(), appID, installID, PrintJSON)
		}),
	}
	installCmd.Flags().StringVar(&installID, "install", "", "The ID of the install you want to use")
	installCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The ID or name of an app to filter installs by")
	configCmd.AddCommand(installCmd)

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

func (c *cli) clearConfig(ctx context.Context) error {
	apiToken := c.cfg.GetString("api_token")

	c.cfg.Set("org_id", "")
	c.cfg.Set("app_id", "")
	c.cfg.Set("install_id", "")

	c.cfg.Set("api_token", apiToken)

	if err := c.cfg.WriteConfig(); err != nil {
		return err
	}

	cmd := &cobra.Command{}
	cmd.Printf("✅ Configuration cleared.\n")

	return nil
}
