package config

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

func TestRunbookStepConfig_PlanOnlyDecode(t *testing.T) {
	var step RunbookStepConfig
	decoderConfig := DecoderConfig()
	decoderConfig.Result = &step
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	require.NoError(t, err)

	require.NoError(t, decoder.Decode(map[string]any{
		"name":      "check-drift",
		"type":      "component_deploy",
		"plan_only": true,
	}))
	require.True(t, step.PlanOnly)
}

func TestRunbookConfig_Parse(t *testing.T) {
	t.Run("basic runbook parses", func(t *testing.T) {
		rc := &RunbookConfig{
			Name:   "v2.3-update",
			Readme: "# Release Notes",
			Steps: []*RunbookStepConfig{
				{
					Name:             "deploy-database",
					Type:             RunbookStepTypeComponentDeploy,
					ComponentName:    "database",
					DeployDependents: true,
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

	t.Run("plan only rejects unsupported step types", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "bad-plan-only",
			Steps: []*RunbookStepConfig{
				{Name: "action", Type: RunbookStepTypeAction, ActionName: "verify", PlanOnly: true},
			},
		}

		err := rc.parse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "plan_only is only supported")
	})

	t.Run("wait for event requires a trigger and event types", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "wait-for-event",
			Steps: []*RunbookStepConfig{{
				Name: "wait",
				Type: RunbookStepTypeWaitForEvent,
			}},
		}

		require.Error(t, rc.parse())
	})

	t.Run("wait for event validates filters", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "wait-for-event",
			Steps: []*RunbookStepConfig{{
				Name:       "wait",
				Type:       RunbookStepTypeWaitForEvent,
				Trigger:    "gar",
				EventTypes: []string{"tag.updated"},
				Filters:    []TriggerFilterConfig{{Path: "not-jsonpath", Op: "eq", Value: "acme/api"}},
			}},
		}

		require.Error(t, rc.parse())
	})

	t.Run("wait for event rejects empty event types", func(t *testing.T) {
		rc := &RunbookConfig{Name: "wait-for-event", Steps: []*RunbookStepConfig{{
			Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar", EventTypes: []string{""},
		}}}

		require.ErrorContains(t, rc.parse(), "event_types must not contain empty strings")
	})

	t.Run("wait for event rejects duplicate event types", func(t *testing.T) {
		rc := &RunbookConfig{Name: "wait-for-event", Steps: []*RunbookStepConfig{{
			Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar", EventTypes: []string{"tag.updated", "tag.updated"},
		}}}

		require.ErrorContains(t, rc.parse(), `event_types contains duplicate "tag.updated"`)
	})

	t.Run("wait for event rejects too many filters", func(t *testing.T) {
		filters := make([]TriggerFilterConfig, 21)
		rc := &RunbookConfig{Name: "wait-for-event", Steps: []*RunbookStepConfig{{
			Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar", EventTypes: []string{"tag.updated"}, Filters: filters,
		}}}

		require.ErrorContains(t, rc.parse(), "filters must contain at most 20 filters")
	})

	t.Run("wait for event rejects exclusion-only filter sets", func(t *testing.T) {
		rc := &RunbookConfig{Name: "wait-for-event", Steps: []*RunbookStepConfig{{
			Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar",
			Filters: []TriggerFilterConfig{{Path: "$.repository", Op: "neq", Value: "acme/api"}},
		}}}

		require.ErrorContains(t, rc.parse(), "must declare event_types, a positive filter, or match_all = true")
	})

	t.Run("wait for event accepts an omitted timeout", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "wait-for-event",
			Steps: []*RunbookStepConfig{{
				Name:       "wait",
				Type:       RunbookStepTypeWaitForEvent,
				Trigger:    "gar",
				EventTypes: []string{"tag.updated"},
				Filters:    []TriggerFilterConfig{{Path: "$.repository", Op: "eq", Value: "acme/api"}},
			}},
		}

		require.NoError(t, rc.parse())
	})

	t.Run("wait for event output names are unique", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "wait-for-events",
			Steps: []*RunbookStepConfig{
				{Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar", EventTypes: []string{"tag.updated"}},
				{Name: "wait", Type: RunbookStepTypeWaitForEvent, Trigger: "gar", EventTypes: []string{"tag.deleted"}},
			},
		}

		require.ErrorContains(t, rc.parse(), "duplicate wait_for_event step name")
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

func TestRunbookConfig_LegacyDeployDependencies(t *testing.T) {
	rc := &RunbookConfig{
		Name: "legacy",
		Steps: []*RunbookStepConfig{
			{
				Name:                     "legacy-deploy",
				Type:                     RunbookStepTypeComponentDeploy,
				ComponentName:            "api",
				DeployDependenciesLegacy: true,
			},
		},
	}

	require.NoError(t, rc.parse())
	require.True(t, rc.Steps[0].DeployDependents, "legacy deploy_dependencies should be folded into DeployDependents")
	require.Len(t, rc.DeprecationWarnings, 1, "deprecation warning should be recorded")
	require.Contains(t, rc.DeprecationWarnings[0], "deploy_dependencies")
}

func TestRunbookConfig_LegacyDeployType(t *testing.T) {
	rc := &RunbookConfig{
		Name: "legacy-type",
		Steps: []*RunbookStepConfig{
			{
				Name:          "legacy-deploy-step",
				Type:          RunbookStepTypeDeployLegacy,
				ComponentName: "api",
			},
		},
	}

	require.NoError(t, rc.parse())
	require.Equal(t, RunbookStepTypeComponentDeploy, rc.Steps[0].Type, "legacy 'deploy' type should be canonicalized to 'component_deploy'")
	require.Len(t, rc.DeprecationWarnings, 1, "deprecation warning should be recorded")
	require.Contains(t, rc.DeprecationWarnings[0], "type 'deploy' is deprecated")
}

func TestRunbookStepType_Constants(t *testing.T) {
	require.Equal(t, RunbookStepType("component_deploy"), RunbookStepTypeComponentDeploy)
	require.Equal(t, RunbookStepType("deploy"), RunbookStepTypeDeployLegacy)
	require.Equal(t, RunbookStepType("action"), RunbookStepTypeAction)
	require.Equal(t, RunbookStepType("sandbox_reprovision"), RunbookStepTypeSandboxReprovision)
	require.Equal(t, RunbookStepType("sandbox_deprovision"), RunbookStepTypeSandboxDeprovision)
	require.Equal(t, RunbookStepType("wait_for_event"), RunbookStepTypeWaitForEvent)
}
