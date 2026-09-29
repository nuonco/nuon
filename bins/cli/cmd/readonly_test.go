package cmd

import "testing"

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

func TestReadOnlyAllowlistHasNoStaleConfigsEntry(t *testing.T) {
	if _, ok := readOnlyCommands["list-configs"]; ok {
		t.Error(`"list-configs" matches no command; the real command is named "configs"`)
	}
}
