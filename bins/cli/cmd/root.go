package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	PrintJSON             bool = false
	ConfigFile            string
	DefaultConfigFilePath string = "~/.nuon"
)

var (
	CoreGroup       = cobra.Group{ID: "core", Title: "Core Commands"}
	InstallGroup    = cobra.Group{ID: "install", Title: "Install Commands"}
	HelpGroup       = cobra.Group{ID: "help", Title: "Help Commands"}
	AdditionalGroup = cobra.Group{ID: "additional", Title: "Additional Commands"}
)

// newRootCmd constructs a new root cobra command, which all other commands will be nested under. If there are any flags
// or other settings that we want to be "global", they should be configured on this command.
func (c *cli) rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nuon",
		Short: "Work with Nuon from the command line.",
		Long:  c.getLongDescription(),
		Example: `nuon auth login
nuon sync
`,
		SilenceUsage:      false,
		SilenceErrors:     false,
		PersistentPreRunE: c.persistentPreRunE,
	}

	rootCmd.PersistentFlags().BoolVarP(&PrintJSON, "json", "j", false, "print output as json")
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "f", DefaultConfigFilePath, "path to custom config file. Can also be set using the NUON_CONFIG_FILE env var.")
	// alias so we can migrate from -f to -C
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config-file", "C", DefaultConfigFilePath, "path to custom config file. Can also be set using the NUON_CONFIG_FILE env var.")

	rootCmd.AddGroup(
		&CoreGroup,
		&InstallGroup,
		&HelpGroup,
		&AdditionalGroup,
	)

	rootCmd.SetCompletionCommandGroupID(HelpGroup.ID)
	rootCmd.SetHelpCommandGroupID(HelpGroup.ID)

	cmds := []*cobra.Command{
		// Core commands
		c.authCmd(),
		c.configCmd(),
		c.appsCmd(),
		c.syncCmd(),

		// Install commands
		c.installsCmd(),

		// Help commands
		c.versionCmd(),
		c.docsCmd(),
		c.exitCodesCmd(),

		// Additional commands
		c.actionsCmd(),
		c.componentsCmd(),
		c.installersCmd(),
		c.orgsCmd(),
		c.secretsCmd(),
		c.buildsCmd(),
		c.devCmd(),
	}

	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd
}

// getLongDescription returns the appropriate authentication status message
func (c *cli) getLongDescription() string {
	status := "Work with Nuon from the command line.\n\n"

	// Try to initialize config if it's not already initialized
	if c.cfg == nil {
		if err := c.initConfig(); err != nil {
			status += "❌ You are not signed-in. Run `nuon auth login` to get started."
			return status
		}
	}

	// If no API token is configured, user is not logged in
	if c.cfg.APIToken == "" {
		status += "❌ You are not signed-in. Run `nuon auth login` to get started."
		return status
	}

	// Try to initialize API client if it's not already initialized
	if c.apiClient == nil {
		if err := c.initAPIClient(); err != nil {
			status += "❌ Unable to connect to Nuon. Run `nuon auth login` to get started."
			return status
		}
	}

	// Try to validate the token by getting current user
	_, err := c.apiClient.GetCurrentUser(context.Background())
	if err != nil {
		status += "❌ Your session has expired. Run `nuon auth login` to sign in again."
		return status
	}

	// User is authenticated.
	status += fmt.Sprintf("✅ You are logged into %s.", c.cfg.APIURL)

	// If an org is already configured, show that.
	orgID := c.cfg.OrgID
	status += "\n\n"
	if orgID != "" {
		status += fmt.Sprintf("org: %s", orgID)
	}

	// Add app info if an app is selected
	appID := c.cfg.GetString("app_id")
	if appID != "" {
		status += "\n"
		status += fmt.Sprintf("app: %s", appID)
	}

	// Add install info if an install is selected
	installID := c.cfg.GetString("install_id")
	if installID != "" {
		status += "\n"
		status += fmt.Sprintf("install: %s", installID)
	}

	return status
}
