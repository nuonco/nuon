package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "nuonctl",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
