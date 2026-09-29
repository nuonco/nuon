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

func (c *cli) rootCmd() *cobra.Command {
	if c.cfg == nil {
		_ = c.initConfig()
	}

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
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config-file", "f", DefaultConfigFilePath, "path to custom config file. Can also be set using the NUON_CONFIG_FILE env var.")
	_ = rootCmd.PersistentFlags().MarkDeprecated("config-file", "use --config/-C instead; -f and --config-file will be removed in a future release")

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
		c.authCmd(),
		c.configCmd(),
		c.appsCmd(),
		c.branchesCmd(),
		c.syncCmd(),

		c.installsCmd(),

		c.versionCmd(),
		c.docsCmd(),
		c.exitCodesCmd(),

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
		c.agentsCmd(),
	}

	if config.Debug() {
		cmds = append(cmds, c.debugCmd())
	}

	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	extMgr := extensions.New(extensionsDir())
	if exts, err := extMgr.List(); err == nil {
		for _, ext := range exts {
			rootCmd.AddCommand(c.extensionProxyCmd(ext))
		}
	}

	return rootCmd
}

func (c *cli) getLongDescription() string {
	status := "Work with Nuon from the command line.\n\n"

	if c.cfg == nil {
		if err := c.initConfig(); err != nil {
			status += "❌ You are not signed-in. Run `nuon auth login` to get started."
			return status
		}
	}

	if c.cfg.APIToken == "" {
		status += "❌ You are not signed-in. Run `nuon auth login` to get started."
		return status
	}

	if c.apiClient == nil {
		if err := c.initAPIClient(); err != nil {
			status += "❌ Unable to connect to Nuon. Run `nuon auth login` to get started."
			return status
		}
	}

	_, err := c.getCurrentUser(context.Background())
	if err != nil {
		status += styles.TextError.Render("Your session has expired. Run `nuon auth login` to sign in again.")
		return status
	}

	status += fmt.Sprintf("✅ You are logged into %s.", c.cfg.APIURL)

	ctx := context.Background()

	orgID := c.cfg.OrgID
	status += "\n\n"
	if orgID != "" {
		if org, err := c.apiClient.GetOrg(ctx); err == nil && org != nil && org.Name != "" {
			status += fmt.Sprintf("org: %s (%s)", org.Name, orgID)
		} else {
			status += fmt.Sprintf("org: %s", orgID)
		}
	}

	appID := c.cfg.GetString("app_id")
	if appID != "" {
		status += "\n"
		if app, err := c.apiClient.GetApp(ctx, appID); err == nil && app != nil && app.Name != "" {
			status += fmt.Sprintf("app: %s (%s)", app.Name, appID)
		} else {
			status += fmt.Sprintf("app: %s", appID)
		}
	}

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
