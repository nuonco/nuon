package cmd

import "testing"

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
		"runs",
	}

	for _, name := range readOnlyLeafCommands {
		if _, ok := readOnlyCommands[name]; !ok {
			t.Errorf("read-only mode should allow %q, but it is not in readOnlyCommands", name)
		}
	}
}

// TestReadOnlyAllowlistHasNoStaleConfigsEntry ensures the dead "list-configs"
// entry does not creep back in place of the real "configs" command.
func TestReadOnlyAllowlistHasNoStaleConfigsEntry(t *testing.T) {
	if _, ok := readOnlyCommands["list-configs"]; ok {
		t.Error(`"list-configs" matches no command; the real command is named "configs"`)
	}
}
