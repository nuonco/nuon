package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// collectCommandNames walks the command tree and returns the set of every
// command name and alias.
func collectCommandNames(cmd *cobra.Command, out map[string]struct{}) {
	out[cmd.Name()] = struct{}{}
	for _, a := range cmd.Aliases {
		out[a] = struct{}{}
	}
	for _, sub := range cmd.Commands() {
		collectCommandNames(sub, out)
	}
}

// TestReadOnlyAllowlistEntriesMatchRealCommands guards against stale allowlist
// entries: every name in readOnlyCommands must correspond to an actual command
// in the tree, otherwise it silently protects nothing (and the real command
// stays blocked). A few names are intentionally not command names.
func TestReadOnlyAllowlistEntriesMatchRealCommands(t *testing.T) {
	names := map[string]struct{}{}
	collectCommandNames((&cli{}).rootCmd(), names)

	// help/completion are cobra built-ins added lazily; init has no command.
	exempt := map[string]struct{}{
		"help":       {},
		"completion": {},
		"init":       {},
	}

	for name := range readOnlyCommands {
		if _, ok := exempt[name]; ok {
			continue
		}
		if _, ok := names[name]; !ok {
			t.Errorf("readOnlyCommands has %q but no command in the tree is named that; either it is stale or the command was renamed", name)
		}
	}
}

// TestReadOnlyAllowsConfigViewers guards against read-only mode blocking
// non-mutating commands. These leaf commands only read remote state, so they
// must be permitted when --read-only / NUON_READ_ONLY is set. `apps configs`
// regressed here: the allowlist held a stale "list-configs" name that matched
// no command, so `nuon apps configs` was blocked.
func TestReadOnlyAllowsConfigViewers(t *testing.T) {
	readOnlyLeafCommands := []string{
		"configs",
		"sandbox-config",
		"input-config",
		"runner-config",
		"plan",
		"api-token",
		"runs",
	}

	for _, name := range readOnlyLeafCommands {
		if _, ok := readOnlyCommands[name]; !ok {
			t.Errorf("read-only mode should allow %q, but it is not in readOnlyCommands", name)
		}
	}
}
