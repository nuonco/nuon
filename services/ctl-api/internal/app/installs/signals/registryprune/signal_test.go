package registryprune

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
)

func TestRunMode(t *testing.T) {
	tests := map[string]struct {
		mode        string
		forceDryRun bool
		wantRun     bool
		wantDryRun  bool
	}{
		"off skips":                       {mode: config.RegistryRetentionModeOff},
		"off with admin dry run previews": {mode: config.RegistryRetentionModeOff, forceDryRun: true, wantRun: true, wantDryRun: true},
		"dry_run reports only":            {mode: config.RegistryRetentionModeDryRun, wantRun: true, wantDryRun: true},
		"enforce deletes":                 {mode: config.RegistryRetentionModeEnforce, wantRun: true},
		"enforce with admin dry run":      {mode: config.RegistryRetentionModeEnforce, forceDryRun: true, wantRun: true, wantDryRun: true},
		"unknown mode never deletes":      {mode: "bogus", wantRun: true, wantDryRun: true},
		"empty mode never deletes":        {mode: "", wantRun: true, wantDryRun: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			run, dryRun := runMode(tc.mode, tc.forceDryRun)
			assert.Equal(t, tc.wantRun, run)
			assert.Equal(t, tc.wantDryRun, dryRun)
		})
	}
}
