package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/extensions"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var (
	PrintJSON             bool = false
	Output                string
	ReadOnly              bool = false
	Debug                 bool = false
	ConfigFile            string
	DefaultConfigFilePath string = "~/.nuon"
)

var (
	CoreGroup       = cobra.Group{ID: "core", Title: "Core Commands"}
	InstallGroup    = cobra.Group{ID: "install", Title: "Install Commands"}
	HelpGroup       = cobra.Group{ID: "help", Title: "Help Commands"}
	AdditionalGroup = cobra.Group{ID: "additional", Title: "Additional Commands"}
	ExtensionGroup  = cobra.Group{ID: "extensions", Title: "Extensions"}
)

// newRootCmd constructs a new root cobra command, which all other commands will be nested under. If there are any flags
// or other settings that we want to be "global", they should be configured on this command.
func (c *cli) rootCmd() *cobra.Command {
	// Commands read cfg while being constructed (cfg.Preview gates required flags).
	if c.cfg == nil {
		_ = c.initConfig()
	}

	// Long is set by populateRootLongDescription: building it calls the API.
	rootCmd := &cobra.Command{
		Use:   "nuon",
		Short: "Work with Nuon from the command line.",
		Example: `nuon auth login
nuon sync
`,
		SilenceUsage:      false,
		SilenceErrors:     false,
		PersistentPreRunE: c.persistentPreRunE,
	}

	rootCmd.PersistentFlags().BoolVarP(&PrintJSON, "json", "j", false, "print output as json (shorthand for --output json)")
	_ = rootCmd.PersistentFlags().MarkDeprecated("json", "use --output json instead; --json will be removed in a future release")
	rootCmd.PersistentFlags().StringVar(&Output, "output", "table", "output format: table, json, or agent. 'agent' is machine-friendly for LLM/agent use (non-interactive, results wrapped in a stable {ok,data,error} envelope on stdout). Can also be set with NUON_OUTPUT.")
	rootCmd.PersistentFlags().BoolVar(&ReadOnly, "read-only", false, "block commands that modify state; safe default when driving the CLI with an agent. Can also be set with NUON_READ_ONLY=1.")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "print per-request API timing (DNS, connect, TLS, server, transfer) to stderr")
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "C", DefaultConfigFilePath, "path to custom config file. Can also be set using the NUON_CONFIG_FILE env var.")

	rootCmd.AddGroup(
		&CoreGroup,
		&InstallGroup,
		&HelpGroup,
		&AdditionalGroup,
		&ExtensionGroup,
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
		c.orgsCmd(),
		c.serviceAccountsCmd(),
		c.rolesCmd(),
		c.secretsCmd(),
		c.buildsCmd(),
		c.loginCmd(),
		c.extensionsCmd(),
		c.runbooksCmd(),
		c.triggersCmd(),
		c.mcpCmd(),
	}

	if config.Debug() {
		cmds = append(cmds, c.debugCmd())
	}

	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	// Register installed extensions as top-level proxy commands.
	extMgr := extensions.New(extensionsDir())
	if exts, err := extMgr.List(); err == nil {
		for _, ext := range exts {
			rootCmd.AddCommand(c.extensionProxyCmd(ext))
		}
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
	_, err := c.getCurrentUser(context.Background())
	if err != nil {
		status += styles.TextError.Render("Your session has expired. Run `nuon auth login` to sign in again.")
		return status
	}

	// User is authenticated.
	status += fmt.Sprintf("✅ You are logged into %s.", c.cfg.APIURL)

	ctx := context.Background()

	// If an org is already configured, show that.
	orgID := c.cfg.OrgID
	status += "\n\n"
	if orgID != "" {
		if org, err := c.apiClient.GetOrg(ctx); err == nil && org != nil && org.Name != "" {
			status += fmt.Sprintf("org: %s (%s)", org.Name, orgID)
		} else {
			status += fmt.Sprintf("org: %s", orgID)
		}
	}

	// Add app info if an app is selected
	appID := c.cfg.GetString("app_id")
	if appID != "" {
		status += "\n"
		if app, err := c.apiClient.GetApp(ctx, appID); err == nil && app != nil && app.Name != "" {
			status += fmt.Sprintf("app: %s (%s)", app.Name, appID)
		} else {
			status += fmt.Sprintf("app: %s", appID)
		}
	}

	// Add install info if an install is selected
	installID := c.cfg.GetString("install_id")
	if installID != "" {
		status += "\n"
		if install, err := c.apiClient.GetInstall(ctx, installID); err == nil && install != nil && install.Name != "" {
			status += fmt.Sprintf("install: %s (%s)", install.Name, installID)
		} else {
			status += fmt.Sprintf("install: %s", installID)
		}
	}

	return status
}
