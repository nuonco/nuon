package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{}

var (
	logger, _ = zap.NewProduction(zap.Fields(zap.String("type", "api")))
)

//nolint:gochecknoinits
func init() {
	flags := rootCmd.Flags()

	flags.String("service_name", "api", "the name of the service")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
