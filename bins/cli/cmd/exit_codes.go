package cmd

import (
	"github.com/spf13/cobra"
)

func (c *cli) exitCodesCmd() *cobra.Command {
	exitCodesCmd := &cobra.Command{
		Use:               "exit-codes",
		Short:             "Learn about exit codes",
		PersistentPreRunE: c.persistentPreRunE,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Exit codes:\n")
			cmd.Printf("  0 - Success\n")
			cmd.Printf("  1 - General error\n")
			cmd.Printf("  2 - CLI initialization or execution error\n")
		},
		GroupID: HelpGroup.ID,
	}

	return exitCodesCmd
}

