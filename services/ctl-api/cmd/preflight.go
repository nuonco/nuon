package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/preflight"
)

func (c *cli) registerPreflight() {
	cmd := &cobra.Command{
		Use:   "preflight [checks...]",
		Short: "validate configuration and service connectivity",
		Long: `Run preflight checks to validate that required configuration is present
and that external services (database, temporal, auth providers, etc.) are reachable.

With no arguments, all checks are run. Pass check names to run a subset.

Available checks: rds, clickhouse, temporal, nuon-auth, auth0, github, aws`,
		Run: runPreflight,
	}
	rootCmd.AddCommand(cmd)
}

func runPreflight(_ *cobra.Command, args []string) {
	cfg, err := internal.NewPreflightConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to load config: %v\n", err)
		os.Exit(2)
	}

	results := preflight.Run(cfg, args)
	code := preflight.PrintResults(results)
	os.Exit(code)
}
