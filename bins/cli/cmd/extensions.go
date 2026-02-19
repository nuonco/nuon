package cmd

import (
	"errors"
	"fmt"
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
		Use:     "extensions",
		Short:   "Manage CLI extensions [preview]",
		Aliases: []string{"ext"},
		GroupID: AdditionalGroup.ID,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := c.persistentPreRunE(cmd, args); err != nil {
				return err
			}
			if !c.cfg.Preview {
				return errors.New("[NUON_PREVIEW=false] extensions are a preview feature, set NUON_PREVIEW=true to enable")
			}
			return nil
		},
		Annotations: skipAuthAnnotation(),
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
		Use:         "install <repo>",
		Short:       "Install an extension",
		Long:        "Install an extension from a GitHub repository. Accepts full repo (org/nuon-ext-name) or shorthand (name).",
		Args:        cobra.ExactArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			if err := mgr.EnsureDir(); err != nil {
				return err
			}

			spinner := ui.NewSpinnerView(PrintJSON)
			spinner.Start(fmt.Sprintf("Installing extension %s...", args[0]))

			ext, err := mgr.Install(args[0])
			if err != nil {
				spinner.Fail(err)
				return err
			}

			spinner.Success(fmt.Sprintf("Installed %s %s", ext.Name, ext.Version))

			if reservedCommandNames[ext.Name] {
				ui.PrintWarning(fmt.Sprintf("Warning: extension %q conflicts with a built-in command. Use `nuon ext exec %s` to run it.", ext.Name, ext.Name))
			}

			if PrintJSON {
				ui.PrintJSON(ext)
			}
			return nil
		}),
	}
}

func (c *cli) extUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "upgrade [name]",
		Short:       "Upgrade extensions",
		Long:        "Upgrade a specific extension or all installed extensions.",
		Args:        cobra.MaximumNArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())

			if len(args) == 0 {
				spinner := ui.NewSpinnerView(PrintJSON)
				spinner.Start("Upgrading all extensions...")

				results, err := mgr.UpgradeAll()
				if err != nil {
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
						fmt.Printf("  %s: %s\n", r.Name, r.Error)
					} else if r.OldVersion != r.NewVersion {
						fmt.Printf("  %s: %s -> %s\n", r.Name, r.OldVersion, r.NewVersion)
					} else {
						fmt.Printf("  %s: already up to date (%s)\n", r.Name, r.OldVersion)
					}
				}
				return nil
			}

			spinner := ui.NewSpinnerView(PrintJSON)
			spinner.Start(fmt.Sprintf("Upgrading %s...", args[0]))

			if err := mgr.Upgrade(args[0]); err != nil {
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
				return err
			}
			view := ui.NewListView()
			view.Print(fmt.Sprintf("Removed extension %s", args[0]))
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

			spinner := ui.NewSpinnerView(PrintJSON)
			spinner.Start("Searching for extensions...")

			exts, err := mgr.Browse(org)
			if err != nil {
				spinner.Fail(err)
				return err
			}

			spinner.Success(fmt.Sprintf("Found %d extension(s)", len(exts)))

			if PrintJSON {
				ui.PrintJSON(exts)
				return nil
			}

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
		Annotations:        skipAuthAnnotation(),
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
		if c.cfg.APIURL != "" {
			env["NUON_API_URL"] = c.cfg.APIURL
		}
		if c.cfg.OrgID != "" {
			env["NUON_ORG_ID"] = c.cfg.OrgID
		}
		if c.cfg.AppID != "" {
			env["NUON_APP_ID"] = c.cfg.AppID
		}
		if c.cfg.InstallID != "" {
			env["NUON_INSTALL_ID"] = c.cfg.InstallID
		}
		if c.cfg.APIToken != "" {
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
		Annotations:        skipAuthAnnotation(),
		PersistentPreRunE:  c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			return mgr.Exec(ext.Name, args, c.extensionEnv())
		}),
	}
}
