package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

const readOnlyEnvVar = "NUON_READ_ONLY"

// readOnlyCommands are leaf commands allowed in read-only mode: anything that
// does not mutate remote state (local config selection and scaffolding are
// allowed). Default-deny: new read commands must be added here explicitly.
var readOnlyCommands = map[string]struct{}{
	"list":                 {},
	"get":                  {},
	"current":              {},
	"current-inputs":       {},
	"latest":               {},
	"latest-config":        {},
	"configs":              {},
	"sandbox-config":       {},
	"input-config":         {},
	"runner-config":        {},
	"list-deploys":         {},
	"list-invites":         {},
	"list-vcs-connections": {},
	"get-deploy":           {},
	"get-run":              {},
	"recent-runs":          {},
	"runs":                 {},
	"plan":                 {},
	"logs":                 {},
	"deploy-logs":          {},
	"sandbox-run-logs":     {},
	"sandbox-runs":         {},
	"sandbox-outputs":      {},
	"outputs":              {},
	"inputs":               {},
	"steps":                {},
	"components":           {},
	"workflows":            {},
	"watch":                {},
	"validate":             {},
	"print-config":         {},
	"generate-config":      {},
	"id":                   {},
	"browse":               {},
	"docs":                 {},
	"exit-codes":           {},
	"version":              {},
	"help":                 {},
	"completion":           {},
	"select":               {},
	"deselect":             {},
	"unset-current":        {},
	"init":                 {},
	"mcp":                  {},
}

func readOnlyFromEnv() bool {
	v := os.Getenv(readOnlyEnvVar)
	return v == "true" || v == "1"
}

// guardReadOnly blocks commands that mutate remote state when read-only mode
// is on.
func guardReadOnly(cmd *cobra.Command) error {
	if !ReadOnly && !readOnlyFromEnv() {
		return nil
	}
	ReadOnly = true

	if _, ok := readOnlyCommands[cmd.Name()]; ok {
		return nil
	}

	return ui.PrintError(&ui.CLIUserError{
		Msg: fmt.Sprintf("read-only mode: `%s` is disabled because it may modify state. Unset %s or drop --read-only to run it.", cmd.CommandPath(), readOnlyEnvVar),
	})
}
