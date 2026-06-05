package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunbookConfig_Parse(t *testing.T) {
	t.Run("basic runbook parses", func(t *testing.T) {
		rc := &RunbookConfig{
			Name:   "v2.3-update",
			Readme: "# Release Notes",
			Steps: []*RunbookStepConfig{
				{
					Name:               "deploy-database",
					Type:               RunbookStepTypeDeploy,
					ComponentName:      "database",
					DeployDependencies: true,
				},
				{
					Name:       "run-migrations",
					Type:       RunbookStepTypeAction,
					ActionName: "database-migration",
				},
				{
					Name:           "post-validation",
					Type:           RunbookStepTypeAction,
					Command:        "./validate.sh",
					InlineContents: "#!/bin/sh\ncurl -sf https://api.example.com/health",
					Timeout:        "2m",
					EnvVarMap:      map[string]string{"API_URL": "https://example.com"},
				},
				{Name: "sbx-reprovision", Type: RunbookStepTypeSandboxReprovision, Role: "custom-role"},
				{Name: "sbx-deprovision", Type: RunbookStepTypeSandboxDeprovision},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
	})

	t.Run("invalid timeout returns error", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "bad-timeout",
			Steps: []*RunbookStepConfig{
				{
					Name:    "step1",
					Type:    RunbookStepTypeAction,
					Command: "echo hello",
					Timeout: "not-a-duration",
				},
			},
		}

		err := rc.parse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("nil runbook parses", func(t *testing.T) {
		var rc *RunbookConfig
		err := rc.parse()
		require.NoError(t, err)
	})

	t.Run("template refs extract dependencies", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "with-refs",
			Steps: []*RunbookStepConfig{
				{
					Name:    "step1",
					Type:    RunbookStepTypeAction,
					Command: "curl {{.component.api.endpoint}}/health",
				},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
		// Dependencies should be extracted from template references
		// (depends on refs.Parse implementation)
	})
}

func TestRunbookStepType_Constants(t *testing.T) {
	require.Equal(t, RunbookStepType("deploy"), RunbookStepTypeDeploy)
	require.Equal(t, RunbookStepType("action"), RunbookStepTypeAction)
	require.Equal(t, RunbookStepType("sandbox_reprovision"), RunbookStepTypeSandboxReprovision)
	require.Equal(t, RunbookStepType("sandbox_deprovision"), RunbookStepTypeSandboxDeprovision)
}
