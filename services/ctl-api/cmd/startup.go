package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *cli) registerStartup() error {
	var runStartupCmd = &cobra.Command{
		Use:   "startup",
		Short: "startup hook that is run on deploy",
		Run:   c.runStartup,
	}
	rootCmd.AddCommand(runStartupCmd)
	return nil
}

func (c *cli) runStartup(cmd *cobra.Command, _ []string) {
	fmt.Println("hello world")
}
