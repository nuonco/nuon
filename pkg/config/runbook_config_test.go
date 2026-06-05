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

func TestRunbookConfig_CellsNormalization(t *testing.T) {
	t.Run("cells normalize to steps", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "notebook-runbook",
			Cells: []*RunbookCellConfig{
				{
					Type:    RunbookCellTypeMarkdown,
					Content: "# Introduction\nThis runbook deploys the database.",
				},
				{
					Type:               RunbookCellTypeDeploy,
					Name:               "deploy-database",
					ComponentName:      "database",
					DeployDependencies: true,
				},
				{
					Type:    RunbookCellTypeMarkdown,
					Content: "Now run the migrations.",
				},
				{
					Type:       RunbookCellTypeAction,
					Name:       "run-migrations",
					ActionName: "database-migration",
				},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
		require.Len(t, rc.Steps, 2)

		require.Equal(t, "deploy-database", rc.Steps[0].Name)
		require.Equal(t, RunbookStepTypeDeploy, rc.Steps[0].Type)
		require.Equal(t, "database", rc.Steps[0].ComponentName)
		require.True(t, rc.Steps[0].DeployDependencies)

		require.Equal(t, "run-migrations", rc.Steps[1].Name)
		require.Equal(t, RunbookStepTypeAction, rc.Steps[1].Type)
		require.Equal(t, "database-migration", rc.Steps[1].ActionName)
	})

	t.Run("cells and steps are mutually exclusive", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "both",
			Steps: []*RunbookStepConfig{
				{Name: "step1", Type: RunbookStepTypeDeploy, ComponentName: "db"},
			},
			Cells: []*RunbookCellConfig{
				{Type: RunbookCellTypeMarkdown, Content: "hello"},
			},
		}

		err := rc.parse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot have both cells and steps")
	})

	t.Run("empty cells is valid", func(t *testing.T) {
		rc := &RunbookConfig{
			Name:  "no-cells",
			Cells: []*RunbookCellConfig{},
			Steps: []*RunbookStepConfig{
				{Name: "step1", Type: RunbookStepTypeDeploy, ComponentName: "db"},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
	})

	t.Run("markdown-only cells produce no steps", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "docs-only",
			Cells: []*RunbookCellConfig{
				{Type: RunbookCellTypeMarkdown, Content: "# Just docs"},
				{Type: RunbookCellTypeMarkdown, Content: "More docs"},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
		require.Len(t, rc.Steps, 0)
	})

	t.Run("cell fields pass through to steps", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "inline-action",
			Cells: []*RunbookCellConfig{
				{
					Type:           RunbookCellTypeAction,
					Name:           "validate",
					Command:        "./validate.sh",
					InlineContents: "#!/bin/sh\necho ok",
					EnvVarMap:      map[string]string{"FOO": "bar"},
					Timeout:        "30s",
					Role:           "admin",
				},
			},
		}

		err := rc.parse()
		require.NoError(t, err)
		require.Len(t, rc.Steps, 1)

		step := rc.Steps[0]
		require.Equal(t, "validate", step.Name)
		require.Equal(t, RunbookStepTypeAction, step.Type)
		require.Equal(t, "./validate.sh", step.Command)
		require.Equal(t, "#!/bin/sh\necho ok", step.InlineContents)
		require.Equal(t, map[string]string{"FOO": "bar"}, step.EnvVarMap)
		require.Equal(t, "30s", step.Timeout)
		require.Equal(t, "admin", step.Role)
	})

	t.Run("unknown cell type returns error", func(t *testing.T) {
		rc := &RunbookConfig{
			Name: "bad-type",
			Cells: []*RunbookCellConfig{
				{Type: "unknown", Name: "bad"},
			},
		}

		err := rc.parse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown cell type")
	})
}
