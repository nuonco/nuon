package cmd

import (
	"github.com/spf13/cobra"
)

var (
	validateFlag  bool
	dryRunFlag    bool
	cleanupFlag   bool
	recursiveFlag bool
	importsFlag   bool
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "temporal-gen-v2",
		Short: "Temporal generator v2",
	}

	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newCleanCmd())

	return rootCmd
}
