package cmd

import (
	"github.com/spf13/cobra"
)

// newRootCmd constructs a new root cobra command, which all other commands will be nested under. If there are any flags or other settings that we want to be "global", they should be configured on this command.
func newRootCmd(
	bindConfig bindConfigFunc,
	cmds ...*cobra.Command,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "nuonctl",
		SilenceUsage: true,
		// PersistentPreRunE is only inherited by immediate child commands.
		// We have to copy/paste this on each subcommand, so that it's children will inherit it.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
	}

	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd
}
