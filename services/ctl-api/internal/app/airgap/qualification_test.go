package airgap

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func supportedConfig() *app.AppConfig {
	return &app.AppConfig{
		ID:            "cfg",
		ComponentIDs:  pq.StringArray{"terraform", "helm", "image", "manifest"},
		ActionIDs:     pq.StringArray{"action"},
		SandboxConfig: app.AppSandboxConfig{ID: "sandbox-config", AppConfigID: "cfg", Type: "terraform"},
		ComponentConfigConnections: []app.ComponentConfigConnection{
			{ID: "terraform-config", AppConfigID: "cfg", ComponentID: "terraform", ComponentName: "terraform", Type: app.ComponentTypeTerraformModule, TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{}},
			{ID: "helm-config", AppConfigID: "cfg", ComponentID: "helm", ComponentName: "helm", Type: app.ComponentTypeHelmChart, HelmComponentConfig: &app.HelmComponentConfig{}},
			{ID: "image-config", AppConfigID: "cfg", ComponentID: "image", ComponentName: "image", Type: app.ComponentTypeExternalImage, ExternalImageComponentConfig: &app.ExternalImageComponentConfig{}},
			{ID: "manifest-config", AppConfigID: "cfg", ComponentID: "manifest", ComponentName: "manifest", Type: app.ComponentTypeKubernetesManifest, KubernetesManifestComponentConfig: &app.KubernetesManifestComponentConfig{}},
		},
		ActionWorkflowConfigs: []app.ActionWorkflowConfig{{
			ID: "action-config", AppConfigID: "cfg", ActionWorkflowID: "action", ActionWorkflow: app.ActionWorkflow{Name: "maintenance"},
		}},
	}
}

func TestQualify(t *testing.T) {
	tests := []struct {
		name           string
		platform       string
		mutate         func(*app.AppConfig)
		qualified      bool
		violationCodes []string
		warningCodes   []string
	}{
		{
			name: "supported config passes", platform: "linux/amd64", qualified: true,
			warningCodes: []string{"component.embedded_images_undetected", "component.embedded_images_undetected"},
		},
		{
			name: "inline action is review warning", platform: "linux/amd64", qualified: true,
			mutate: func(c *app.AppConfig) {
				c.ActionWorkflowConfigs[0].Steps = []app.ActionWorkflowStepConfig{{Name: "script", AppConfigID: "cfg", ActionWorkflowConfigID: "action-config", InlineContents: "echo ok"}}
			},
			warningCodes: []string{"action_step.inline_review", "component.embedded_images_undetected", "component.embedded_images_undetected"},
		},
		{
			name: "collects independent failures", platform: "darwin/arm64",
			mutate: func(c *app.AppConfig) {
				c.ComponentConfigConnections[0].AppConfigID = "other"
				c.ComponentConfigConnections[0].ComponentID = "not-a-member"
				c.ComponentConfigConnections[0].Type = app.ComponentTypeDockerBuild
				c.ComponentConfigConnections[0].DockerBuildComponentConfig = &app.DockerBuildComponentConfig{}
				c.SandboxConfig.AppConfigID = "other"
				c.SandboxConfig.Type = "pulumi"
				c.ActionWorkflowConfigs[0].AppConfigID = "other"
				c.ActionWorkflowConfigs[0].Steps = []app.ActionWorkflowStepConfig{{Name: "git", AppConfigID: "other", ActionWorkflowConfigID: "action-config", PublicGitVCSConfig: &app.PublicGitVCSConfig{}}}
				c.StackConfig.CustomNestedStacks = []config.CustomNestedStack{{Name: "asset", TemplateURL: "https://example/{{.asset}}"}}
			},
			violationCodes: []string{
				"action.app_config_mismatch", "action_step.app_config_mismatch",
				"component.config_missing", "component.docker_build_unsupported",
				"component.membership_mismatch", "component.source_config_invalid", "platform.unsupported",
				"sandbox.app_config_mismatch", "sandbox.pulumi_unsupported", "stack_asset.templated_url", "stack_asset.unpinned_custom",
			},
			warningCodes: []string{"action_step.git_excluded", "component.embedded_images_undetected", "component.embedded_images_undetected"},
		},
		{
			name: "cron git action blocks export", platform: "linux/amd64",
			mutate: func(c *app.AppConfig) {
				c.ActionWorkflowConfigs[0].Triggers = []app.ActionWorkflowTriggerConfig{{Type: app.ActionWorkflowTriggerTypeCron, CronSchedule: "0 0 * * *"}}
				c.ActionWorkflowConfigs[0].Steps = []app.ActionWorkflowStepConfig{{Name: "git", AppConfigID: "cfg", ActionWorkflowConfigID: "action-config", PublicGitVCSConfig: &app.PublicGitVCSConfig{}}}
			},
			violationCodes: []string{"action_step.git_unsupported"},
			warningCodes:   []string{"component.embedded_images_undetected", "component.embedded_images_undetected"},
		},
		{
			name: "rejects unsupported component surfaces", platform: "linux/amd64",
			mutate: func(c *app.AppConfig) {
				c.ComponentIDs = pq.StringArray{"job", "pulumi", "unknown"}
				c.ComponentConfigConnections = []app.ComponentConfigConnection{
					{ID: "job-config", AppConfigID: "cfg", ComponentID: "job", Type: app.ComponentTypeJob, JobComponentConfig: &app.JobComponentConfig{}},
					{ID: "pulumi-config", AppConfigID: "cfg", ComponentID: "pulumi", Type: app.ComponentTypePulumi, PulumiComponentConfig: &app.PulumiComponentConfig{}},
					{ID: "unknown-config", AppConfigID: "cfg", ComponentID: "unknown", Type: app.ComponentTypeUnknown},
				}
			},
			violationCodes: []string{"component.job_unsupported", "component.pulumi_unsupported", "component.source_config_invalid", "component.type_unsupported"},
		},
		{
			name: "rejects malformed authoritative membership", platform: "linux/amd64",
			mutate: func(c *app.AppConfig) {
				c.ComponentIDs = append(c.ComponentIDs, "terraform", "")
				c.ActionIDs = append(c.ActionIDs, "action", "")
			},
			violationCodes: []string{"action.member_duplicate", "action.member_invalid", "component.member_duplicate", "component.member_invalid"},
			warningCodes:   []string{"component.embedded_images_undetected", "component.embedded_images_undetected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := supportedConfig()
			if tt.mutate != nil {
				tt.mutate(cfg)
			}
			r := Qualify(cfg, tt.platform)
			require.Equal(t, tt.qualified, r.Qualified)
			require.Equal(t, tt.violationCodes, codes(r.Violations))
			require.Equal(t, tt.warningCodes, codes(r.Warnings))
			assertSorted(t, r.Violations)
			assertSorted(t, r.Warnings)
		})
	}
}

func TestQualifyNilReturnsAllApplicableViolations(t *testing.T) {
	r := Qualify(nil, "windows/amd64")
	require.Equal(t, []string{"app_config.missing", "platform.unsupported"}, codes(r.Violations))
}

func codes(findings []Finding) []string {
	if len(findings) == 0 {
		return nil
	}
	result := make([]string, len(findings))
	for i := range findings {
		result[i] = findings[i].Code
	}
	return result
}

func assertSorted(t *testing.T, findings []Finding) {
	for i := 1; i < len(findings); i++ {
		previous, current := findings[i-1], findings[i]
		require.True(t, previous.Code < current.Code || previous.Code == current.Code && previous.Member <= current.Member)
	}
}
