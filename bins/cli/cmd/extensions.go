package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/extensions"
)

func extensionsDir() string {
	home, _ := homedir.Dir()
	return filepath.Join(home, ".nuon", "extensions")
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
			fmt.Printf("Found %d installed extensions\n", len(exts))
			return nil
		}),
	}
}

func (c *cli) extInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "install <repo>",
		Short:       "Install an extension",
		Long:        "Install an extension from a GitHub repository. Accepts full repo (nuonco/nuon-ext-name) or shorthand (name).",
		Args:        cobra.ExactArgs(1),
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			_, err := mgr.Install(args[0])
			if err != nil {
				return err
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
				_, err := mgr.UpgradeAll()
				return err
			}
			return mgr.Upgrade(args[0])
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
			return mgr.Remove(args[0])
		}),
	}
}

func (c *cli) extBrowseCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "browse",
		Short:       "Browse available extensions",
		Long:        "List available extensions from the nuonco GitHub organization.",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			mgr := extensions.New(extensionsDir())
			exts, err := mgr.Browse()
			if err != nil {
				return err
			}
			fmt.Printf("Found %d available extensions\n", len(exts))
			return nil
		}),
	}
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
			return mgr.Exec(args[0], args[1:], nil)
		}),
	}
}
