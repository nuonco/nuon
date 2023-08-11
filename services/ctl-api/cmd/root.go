package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

//nolint:gochecknoinits
func init() {
	flags := rootCmd.Flags()
	flags.String("service_name", "ctl-api", "the name of the service")
}

func Execute() {
	c := &cli{}
	c.registerAPI()
	c.registerWorker()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
