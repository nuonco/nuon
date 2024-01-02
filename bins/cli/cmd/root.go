package cmd

import (
	"github.com/spf13/cobra"
)

var (
	PrintJSON             bool = false
	ConfigFile            string
	DefaultConfigFilePath string = "~/.nuon"
)

// newRootCmd constructs a new root cobra command, which all other commands will be nested under. If there are any flags
// or other settings that we want to be "global", they should be configured on this command.
func (c *cli) rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "nuon",
		SilenceUsage:      true,
		PersistentPreRunE: c.persistentPreRunE,
	}

	rootCmd.PersistentFlags().BoolVarP(&PrintJSON, "json", "j", false, "print output as json")
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "f", DefaultConfigFilePath, "path to custom config file. Can also be set using the NUON_CONFIG_FILE env var.")

	cmds := []*cobra.Command{
		c.appsCmd(),
		c.buildsCmd(),
		c.componentsCmd(),
		c.installsCmd(),
		c.releasesCmd(),
		c.orgsCmd(),
		c.versionCmd(),
		c.loginCmd(),
	}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd
}
