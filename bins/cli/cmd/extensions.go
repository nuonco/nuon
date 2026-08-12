package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/extensions"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func extensionsDir() string {
	home, _ := homedir.Dir()
	return filepath.Join(home, ".config", "nuon", "extensions")
}

// reservedCommandNames are top-level CLI command names (and aliases) that extensions
// must not shadow. An extension can still be installed, but the user is warned
// that `nuon <name>` will invoke the built-in command, not the extension.
var reservedCommandNames = map[string]bool{
	"auth":       true,
	"config":     true,
	"apps":       true,
	"sync":       true,
	"installs":   true,
	"version":    true,
	"docs":       true,
	"exit-codes": true,
	"actions":    true,
	"components": true,
	"orgs":       true,
	"secrets":    true,
	"builds":     true,
	"dev":        true,
	"login":      true,
	"extensions": true,
	"ext":        true,
	"init":       true,
	"help":       true,
	"completion": true,
}

func (c *cli) extensionsCmd() *cobra.Command {
	extCmd := &cobra.Command{
		Use:               "extensions",
		Short:             "Manage CLI extensions",
		Aliases:           []string{"ext"},
		GroupID:           AdditionalGroup.ID,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
	}

	extCmd.AddCommand(
		c.extListCmd(),
		c.extInstallCmd(),
		c.extUpgradeCmd(),
		c.extRemoveCmd(),
		c.extBrowseCmd(),
		c.extExecCmd(),
	)

	return extCmd
}

func (c *cli) extListCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Aliases:     []string{"ls"},
		Short:       "List installed extensions",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			exts, err := mgr.List()
			if err != nil {
				return err
			}

			if PrintJSON {
				if exts == nil {
					exts = []extensions.InstalledExtension{}
				}
				ui.PrintJSON(exts)
				return nil
			}

			view := ui.NewListView()

			if len(exts) == 0 {
				view.Print("No extensions installed. Run `nuon ext browse` to discover available extensions.")
				return nil
			}

			data := [][]string{
				{"NAME", "VERSION", "REPO", "DESCRIPTION"},
			}
			for _, ext := range exts {
				data = append(data, []string{
					ext.Name,
					ext.Version,
					ext.Repo,
					ext.Description,
				})
			}
			view.Render(data)
			return nil
		}),
	}
}

func (c *cli) extInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <repo-or-path>",
		Short: "Install an extension",
		Long: `Install an extension from a GitHub repository, a local directory, or a local binary.

  GitHub:         nuon ext install api
                  nuon ext install nuonco/nuon-ext-api
                  nuon ext install nuonco/nuon-ext-api@v0.19.798

  Local directory: nuon ext install ./nuon-ext-my-tool
                   (must contain nuon-ext.toml; creates symlink)

  Local binary:   nuon ext install ~/bin/nuon-ext-linter
                  nuon ext install /usr/local/bin/nuon-ext-linter
                  (binary name must use nuon-ext-<name> convention; copies binary)`,
		Args:        cobra.ExactArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			if err := mgr.EnsureDir(); err != nil {
				return ui.PrintError(err)
			}

			var spinner *ui.SpinnerView
			if !PrintJSON {
				spinner = ui.NewSpinnerView(PrintJSON, c.cfg.Interactive)
				spinner.Start(fmt.Sprintf("Installing extension %s...", args[0]))
			}

			ext, err := mgr.Install(args[0])
			if err != nil {
				if PrintJSON {
					return ui.PrintError(err)
				}
				spinner.Fail(err)
				return err
			}

			if reservedCommandNames[ext.Name] {
				ui.PrintWarning(fmt.Sprintf("Warning: extension %q conflicts with a built-in command. Use `nuon ext exec %s` to run it.", ext.Name, ext.Name))
			}

			if PrintJSON {
				ui.PrintJSON(ext)
				return nil
			}

			spinner.Success(fmt.Sprintf("Installed %s %s", ext.Name, ext.Version))
			return nil
		}),
	}
}

func (c *cli) extUpgradeCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:         "upgrade [name]",
		Short:       "Upgrade extensions",
		Long:        "Upgrade a specific extension or all installed extensions. Use --force to re-download even if already at latest version.",
		Args:        cobra.MaximumNArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())

			if len(args) == 0 {
				var spinner *ui.SpinnerView
				if !PrintJSON {
					spinner = ui.NewSpinnerView(PrintJSON, c.cfg.Interactive)
					spinner.Start("Upgrading all extensions...")
				}

				results, err := mgr.UpgradeAll()
				if err != nil {
					if PrintJSON {
						return ui.PrintError(err)
					}
					spinner.Fail(err)
					return err
				}

				if PrintJSON {
					ui.PrintJSON(results)
					return nil
				}

				if len(results) == 0 {
					spinner.Success("No extensions installed")
					return nil
				}

				spinner.Success(fmt.Sprintf("Upgraded %d extension(s)", len(results)))

				for _, r := range results {
					if r.Error != nil {
						ui.Printf("  %s: %s\n", r.Name, r.Error)
					} else if r.OldVersion != r.NewVersion {
						ui.Printf("  %s: %s -> %s\n", r.Name, r.OldVersion, r.NewVersion)
					} else {
						ui.Printf("  %s: already up to date (%s)\n", r.Name, r.OldVersion)
					}
				}
				return nil
			}

			if PrintJSON {
				if err := mgr.Upgrade(args[0], force); err != nil {
					return ui.PrintError(err)
				}
				ext, _ := mgr.Get(args[0])
				ui.PrintJSON(ext)
				return nil
			}

			spinner := ui.NewSpinnerView(PrintJSON, c.cfg.Interactive)
			spinner.Start(fmt.Sprintf("Upgrading %s...", args[0]))

			if err := mgr.Upgrade(args[0], force); err != nil {
				spinner.Fail(err)
				return err
			}

			ext, _ := mgr.Get(args[0])
			if ext != nil {
				spinner.Success(fmt.Sprintf("Upgraded %s to %s", ext.Name, ext.Version))
			} else {
				spinner.Success(fmt.Sprintf("Upgraded %s", args[0]))
			}
			return nil
		}),
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force re-download even if already at latest version")

	return cmd
}

func (c *cli) extRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "remove <name>",
		Short:       "Remove an installed extension",
		Args:        cobra.ExactArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			if err := mgr.Remove(args[0]); err != nil {
				return ui.PrintError(err)
			}
			if PrintJSON {
				ui.PrintJSON(map[string]string{
					"name":    args[0],
					"status":  "removed",
					"message": fmt.Sprintf("Removed extension %s", args[0]),
				})
				return nil
			}
			ui.NewListView().Print(fmt.Sprintf("Removed extension %s", args[0]))
			return nil
		}),
	}
}

func (c *cli) extBrowseCmd() *cobra.Command {
	var org string

	cmd := &cobra.Command{
		Use:         "browse",
		Short:       "Browse available extensions",
		Long:        "List available extensions from a GitHub organization (defaults to nuonco).",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())

			var spinner *ui.SpinnerView
			if !PrintJSON {
				spinner = ui.NewSpinnerView(PrintJSON, c.cfg.Interactive)
				spinner.Start("Searching for extensions...")
			}

			exts, err := mgr.Browse(org)
			if err != nil {
				if PrintJSON {
					return ui.PrintError(err)
				}
				spinner.Fail(err)
				return err
			}

			if PrintJSON {
				ui.PrintJSON(exts)
				return nil
			}

			spinner.Success(fmt.Sprintf("Found %d extension(s)", len(exts)))

			if len(exts) == 0 {
				view := ui.NewListView()
				view.Print("No extensions available.")
				return nil
			}

			view := ui.NewListView()
			data := [][]string{
				{"NAME", "VERSION", "INSTALLED", "REPO", "DESCRIPTION"},
			}
			for _, ext := range exts {
				installed := " "
				if ext.Installed {
					installed = "*"
				}
				data = append(data, []string{
					ext.Name,
					ext.LatestTag,
					installed,
					ext.Repo,
					ext.Description,
				})
			}
			view.Render(data)
			return nil
		}),
	}

	cmd.Flags().StringVar(&org, "org", "", "GitHub organization to browse (default: nuonco)")

	return cmd
}

func (c *cli) extExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "exec <name> [args...]",
		Short:              "Run an extension explicitly",
		Long:               "Run an installed extension by name. Useful if the extension name collides with a built-in command.",
		Args:               cobra.MinimumNArgs(1),
		Annotations:        annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		DisableFlagParsing: true,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			return mgr.Exec(args[0], args[1:], c.extensionEnv())
		}),
	}
}

// extensionEnv builds the environment variables to pass to extensions.
func (c *cli) extensionEnv() map[string]string {
	env := map[string]string{
		"NUON_CONFIG_FILE": ConfigFile,
	}

	if c.cfg != nil {
		if apiURL := c.cfg.GetString("api_url"); apiURL != "" {
			env["NUON_API_URL"] = apiURL
		} else if c.cfg.APIURL != "" {
			env["NUON_API_URL"] = c.cfg.APIURL
		}
		if orgID := c.cfg.GetString("org_id"); orgID != "" {
			env["NUON_ORG_ID"] = orgID
		} else if c.cfg.OrgID != "" {
			env["NUON_ORG_ID"] = c.cfg.OrgID
		}
		if appID := c.cfg.GetString("app_id"); appID != "" {
			env["NUON_APP_ID"] = appID
		} else if c.cfg.AppID != "" {
			env["NUON_APP_ID"] = c.cfg.AppID
		}
		if installID := c.cfg.GetString("install_id"); installID != "" {
			env["NUON_INSTALL_ID"] = installID
		} else if c.cfg.InstallID != "" {
			env["NUON_INSTALL_ID"] = c.cfg.InstallID
		}
		if apiToken := c.cfg.GetString("api_token"); apiToken != "" {
			env["NUON_API_TOKEN"] = apiToken
		} else if c.cfg.APIToken != "" {
			env["NUON_API_TOKEN"] = c.cfg.APIToken
		}
	}

	return env
}

// extensionProxyCmd creates a top-level cobra command that proxies to an installed extension.
func (c *cli) extensionProxyCmd(ext extensions.InstalledExtension) *cobra.Command {
	return &cobra.Command{
		Use:                ext.Name,
		Short:              ext.Description,
		GroupID:            ExtensionGroup.ID,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		Annotations:        annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// With DisableFlagParsing, Cobra skips all flag parsing
			// including the root's persistent flags (e.g. -C ~/.stage).
			// Manually parse them so the config is loaded correctly.
			if preArgs := argsBeforeCommand(ext.Name); len(preArgs) > 0 {
				cmd.Root().PersistentFlags().Parse(preArgs)
			}
			return c.persistentPreRunE(cmd, args)
		},
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			// When DisableFlagParsing is true, Cobra passes all unparsed
			// args including parent persistent flags (e.g. -C ~/.stage)
			// that precede the extension name. Extract only the args that
			// appear after the extension command name in os.Args.
			extArgs := argsAfterCommand(ext.Name)
			return mgr.Exec(ext.Name, extArgs, c.extensionEnv())
		}),
	}
}

// argsBeforeCommand returns the slice of os.Args[1:] that precedes the first
// occurrence of the given command name (i.e. parent flags like "-C ~/.stage").
func argsBeforeCommand(name string) []string {
	for i, arg := range os.Args[1:] {
		if arg == name {
			return os.Args[1 : i+1]
		}
	}
	return nil
}

// argsAfterCommand returns the slice of os.Args that follows the first
// occurrence of the given command name. This ensures that parent flags
// preceding the command (e.g. "-C ~/.stage") are not forwarded.
func argsAfterCommand(name string) []string {
	for i, arg := range os.Args {
		if arg == name {
			return os.Args[i+1:]
		}
	}
	return nil
}
