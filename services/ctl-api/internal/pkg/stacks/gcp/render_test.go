package gcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func testInput() *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "instabcdefghijklmnopqrstuv",
			AppID: "appabcdefghijklmnopqrstuvw",
			OrgID: "orgabcdefghijklmnopqrstuvw",
			GCPAccount: &app.GCPAccount{
				ProjectID: "my-gcp-project",
				Region:    "us-central1",
			},
		},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL: "https://example.com/phone-home",
		},
		InstallState:                 &state.State{},
		AppCfg:                       &app.AppConfig{},
		Runner:                       &app.Runner{ID: "runnerabcdefghijklmnopqrstu", OrgID: "orgabcdefghijklmnopqrstuvw"},
		Settings:                     &app.RunnerGroupSettings{RunnerAPIURL: "https://runner.nuon.co"},
		APIToken:                     "test-token",
		RunnerInitScriptURL:          "https://example.com/init.sh",
		PhonehomeScript:              "echo done",
		VPCNestedStackTemplateURL:    "https://example.com/vpc.yaml",
		RunnerNestedStackTemplateURL: "https://example.com/runner.yaml",
	}
}

func testInputWithBreakGlass() *stacks.TemplateInput {
	inp := testInput()
	inp.AppCfg.BreakGlassConfig = app.AppBreakGlassConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				CloudPlatform: "gcp",
				Type:          app.AWSIAMRoleTypeBreakGlass,
				Policies: []app.AppAWSIAMPolicyConfig{
					{GCPPermissions: []string{"iam.roles.get"}},
				},
			},
		},
	}
	return inp
}

func TestRenderValidJSON(t *testing.T) {
	out, checksum, err := Render(testInput())
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(out, &parsed), "rendered template must be valid JSON")
}

func TestRenderCustomRoles(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		customRoles := resources["google_project_iam_custom_role"].(map[string]any)

		for _, role := range []string{"provision", "maintenance", "deprovision"} {
			assert.Contains(t, customRoles, role)
		}
		assert.NotContains(t, customRoles, "break_glass")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		customRoles := resources["google_project_iam_custom_role"].(map[string]any)

		for _, role := range []string{"provision", "maintenance", "deprovision", "break_glass"} {
			assert.Contains(t, customRoles, role)
		}
	})
}

func TestRenderServiceAccounts(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		sas := resources["google_service_account"].(map[string]any)

		for _, sa := range []string{"runner", "provision", "maintenance", "deprovision"} {
			assert.Contains(t, sas, sa)
		}
		assert.NotContains(t, sas, "break_glass")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		sas := resources["google_service_account"].(map[string]any)

		for _, sa := range []string{"runner", "provision", "maintenance", "deprovision", "break_glass"} {
			assert.Contains(t, sas, sa)
		}
	})
}

func TestRenderTokenCreatorGrants(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		iamMembers := resources["google_service_account_iam_member"].(map[string]any)

		for _, grant := range []string{"provision_token_creator", "maintenance_token_creator", "deprovision_token_creator"} {
			assert.Contains(t, iamMembers, grant)
		}
		assert.NotContains(t, iamMembers, "break_glass_token_creator")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		resources := parsed["resource"].(map[string]any)
		iamMembers := resources["google_service_account_iam_member"].(map[string]any)

		for _, grant := range []string{"provision_token_creator", "maintenance_token_creator", "deprovision_token_creator", "break_glass_token_creator"} {
			assert.Contains(t, iamMembers, grant)
		}
	})
}

func TestRenderOutputs(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		outputs := parsed["output"].(map[string]any)

		for _, output := range []string{"provision_sa_email", "maintenance_sa_email", "deprovision_sa_email", "runner_service_account_email"} {
			assert.Contains(t, outputs, output)
		}
		assert.NotContains(t, outputs, "break_glass_sa_email")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(out, &parsed))

		outputs := parsed["output"].(map[string]any)

		for _, output := range []string{"provision_sa_email", "maintenance_sa_email", "deprovision_sa_email", "break_glass_sa_email", "runner_service_account_email"} {
			assert.Contains(t, outputs, output)
		}
	})
}
