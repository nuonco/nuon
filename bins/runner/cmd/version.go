package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/runner/internal/version"
)

func (c *cli) registerVersion() error {
	rootCmd.AddCommand(&cobra.Command{
		Use:  "version",
		Long: "emit the runner version",
		Run:  c.runVersion,
	})
	return nil
}

func (c *cli) runVersion(cmd *cobra.Command, _ []string) {
	fmt.Printf("%s\n", version.Version)
}
