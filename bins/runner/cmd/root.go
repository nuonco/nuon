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
	c.registerAirgap()

	cmd, err := rootCmd.ExecuteC()
	if err != nil && cmd != nil && cmd.Name() == "airgap" {
		os.Exit(1)
	}
	if err != nil {
		os.Exit(2)
	}
}
