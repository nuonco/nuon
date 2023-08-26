package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Execute() {
	c := &cli{}
	c.registerAPI()
	c.registerWorker()
	c.registerStartup()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
