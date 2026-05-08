package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Execute() {
	c := &cli{}
	c.registerAPI()
	c.registerPublicAPI()
	c.registerInternalAPI()
	c.registerRunnerAPI()
	c.registerAuthAPI()
	c.registerAdminDashboardAPI()
	c.registerSlackAPI()
	c.registerWorker()
	c.registerStartup()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
