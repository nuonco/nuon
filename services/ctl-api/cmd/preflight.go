package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/services/ctl-api/internal/preflight"
)

func (c *cli) registerPreflight() {
	cmd := &cobra.Command{
		Use:   "preflight [checks...]",
		Short: "validate configuration and service connectivity",
		Long: `Run preflight checks to validate that required configuration is present
and that external services (database, temporal, auth providers, etc.) are reachable.

With no arguments, all checks are run. Pass check names to run a subset.
Checks that do not apply to this deployment report as skipped.

--json emits a machine-readable report on stdout, for both runs and --list.
Secret values are never included: each field reports only whether it is set.

Available checks: ` + strings.Join(preflight.Names(), ", "),
		Run: runPreflight,
	}
	cmd.Flags().Bool("list", false, "list checks and the config they read, without connecting")
	cmd.Flags().Bool("json", false, "emit JSON instead of a table")
	rootCmd.AddCommand(cmd)
}

func runPreflight(cmd *cobra.Command, args []string) {
	// Deliberately not internal.NewConfig: it validates the whole struct up
	// front, so a single missing field would abort before any check could
	// report which ones are actually wrong.
	cfg, err := preflight.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	asJSON, _ := cmd.Flags().GetBool("json")

	if list, _ := cmd.Flags().GetBool("list"); list {
		described := preflight.Describe(cfg, args)
		if asJSON {
			if err := preflight.WriteJSONChecks(os.Stdout, described); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(2)
			}

			return
		}

		preflight.PrintChecks(os.Stdout, described)

		return
	}

	results := preflight.Run(context.Background(), cfg, args)
	if asJSON {
		os.Exit(preflight.WriteJSONResults(os.Stdout, results))
	}

	os.Exit(preflight.PrintResults(os.Stdout, results))
}
