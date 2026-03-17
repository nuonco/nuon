package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestResolveStepConfig(t *testing.T) {
	tests := []struct {
		name                   string
		configStepCfg          *models.AppActionWorkflowStepConfig
		step                   *models.AppInstallActionWorkflowRunStep
		stepPlan               *plantypes.ActionWorkflowRunStepPlan
		expectedCommand        string
		expectedInlineContents string
		expectedName           string
		expectedEnvVars        map[string]string
	}{
		{
			name: "regular action uses interpolated inline contents from plan",
			configStepCfg: &models.AppActionWorkflowStepConfig{
				Name:           "run",
				InlineContents: "#!/bin/bash\necho {{ .nuon.install.id }}",
				Command:        "",
			},
			step: &models.AppInstallActionWorkflowRunStep{},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedInlineContents: "#!/bin/bash\necho inst_abc123",
				InterpolatedCommand:        "",
			},
			expectedCommand:        "",
			expectedInlineContents: "#!/bin/bash\necho inst_abc123",
			expectedName:           "run",
		},
		{
			name: "regular action uses interpolated command from plan",
			configStepCfg: &models.AppActionWorkflowStepConfig{
				Name:    "deploy",
				Command: "deploy --id={{ .nuon.install.id }}",
			},
			step: &models.AppInstallActionWorkflowRunStep{},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedCommand: "deploy --id=inst_abc123",
			},
			expectedCommand:        "deploy --id=inst_abc123",
			expectedInlineContents: "",
			expectedName:           "deploy",
		},
		{
			name: "regular action keeps raw config when plan has no interpolated values",
			configStepCfg: &models.AppActionWorkflowStepConfig{
				Name:           "static-step",
				InlineContents: "#!/bin/bash\necho hello",
				Command:        "",
			},
			step: &models.AppInstallActionWorkflowRunStep{},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedInlineContents: "",
				InterpolatedCommand:        "",
			},
			expectedCommand:        "",
			expectedInlineContents: "#!/bin/bash\necho hello",
			expectedName:           "static-step",
		},
		{
			name:          "adhoc action uses interpolated inline contents from plan",
			configStepCfg: nil,
			step: &models.AppInstallActionWorkflowRunStep{
				AdhocConfig: &models.AppAdHocStepConfig{
					Name:           "adhoc-run",
					InlineContents: "#!/bin/bash\necho {{ .nuon.install.id }}",
					Command:        "",
					EnvVars:        map[string]string{"FOO": "bar"},
				},
			},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedInlineContents: "#!/bin/bash\necho inst_abc123",
			},
			expectedCommand:        "",
			expectedInlineContents: "#!/bin/bash\necho inst_abc123",
			expectedName:           "adhoc-run",
			expectedEnvVars:        map[string]string{"FOO": "bar"},
		},
		{
			name:          "adhoc action uses interpolated command from plan",
			configStepCfg: nil,
			step: &models.AppInstallActionWorkflowRunStep{
				AdhocConfig: &models.AppAdHocStepConfig{
					Name:    "adhoc-deploy",
					Command: "deploy --id={{ .nuon.install.id }}",
				},
			},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedCommand: "deploy --id=inst_abc123",
			},
			expectedCommand: "deploy --id=inst_abc123",
			expectedName:    "adhoc-deploy",
		},
		{
			name:          "adhoc action falls back to raw config when plan has no interpolated values",
			configStepCfg: nil,
			step: &models.AppInstallActionWorkflowRunStep{
				AdhocConfig: &models.AppAdHocStepConfig{
					Name:           "adhoc-static",
					InlineContents: "#!/bin/bash\necho hello",
					Command:        "run-something",
				},
			},
			stepPlan: &plantypes.ActionWorkflowRunStepPlan{
				InterpolatedInlineContents: "",
				InterpolatedCommand:        "",
			},
			expectedCommand:        "run-something",
			expectedInlineContents: "#!/bin/bash\necho hello",
			expectedName:           "adhoc-static",
		},
		{
			name:          "no config and no adhoc config returns nil with default name",
			configStepCfg: nil,
			step:          &models.AppInstallActionWorkflowRunStep{},
			stepPlan:      &plantypes.ActionWorkflowRunStepPlan{},
			expectedName:  "adhoc step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, name := resolveStepConfig(tt.configStepCfg, tt.step, tt.stepPlan)
			assert.Equal(t, tt.expectedName, name)

			if tt.expectedCommand == "" && tt.expectedInlineContents == "" && tt.configStepCfg == nil && (tt.step.AdhocConfig == nil) {
				require.Nil(t, cfg)
				return
			}

			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedCommand, cfg.Command)
			assert.Equal(t, tt.expectedInlineContents, cfg.InlineContents)

			if tt.expectedEnvVars != nil {
				assert.Equal(t, tt.expectedEnvVars, cfg.EnvVars)
			}
		})
	}
}
