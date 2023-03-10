package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "run",
}

//nolint:gochecknoinits
func init() {
	flags := rootCmd.Flags()

	flags.String("service_name", "workers-orgs", "the name of the service")
	flags.String("temporal_host", "", "the temporal host and port")
	flags.String("temporal_namespace", "", "the temporal namespace")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
