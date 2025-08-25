package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Execute() {
	c := &cli{}
	c.registerMng()
	c.registerBuild()
	c.registerInstall()
	c.registerRun()
	c.registerVersion()
	c.registerRunLocal()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
