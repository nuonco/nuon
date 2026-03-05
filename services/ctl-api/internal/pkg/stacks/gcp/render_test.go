package gcp

import (
	"encoding/json"
	"strings"
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

// extractTfvars parses the JSON envelope and returns the tfvars string.
func extractTfvars(t *testing.T, out []byte) string {
	t.Helper()
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(out, &envelope))
	tfvars, ok := envelope["tfvars"]
	require.True(t, ok, "envelope must contain 'tfvars' key")
	return tfvars
}

func TestRenderValidJSON(t *testing.T) {
	out, checksum, err := Render(testInput())
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(out, &parsed), "rendered template must be valid JSON")
}

func TestRenderStandardVars(t *testing.T) {
	out, _, err := Render(testInput())
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	expected := map[string]string{
		"nuon_install_id":        `"instabcdefghijklmnopqrstuv"`,
		"nuon_org_id":            `"orgabcdefghijklmnopqrstuvw"`,
		"nuon_app_id":            `"appabcdefghijklmnopqrstuvw"`,
		"runner_api_url":         `"https://runner.nuon.co"`,
		"runner_api_token":       `"test-token"`,
		"runner_id":              `"runnerabcdefghijklmnopqrstu"`,
		"runner_init_script_url": `"https://example.com/init.sh"`,
		"phone_home_url":         `"https://example.com/phone-home"`,
	}
	for key, val := range expected {
		assert.Contains(t, tfvars, key+" ", "tfvars should contain %s", key)
		assert.Contains(t, tfvars, val, "tfvars should contain value %s for %s", val, key)
	}
}

func TestRenderPermissions(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_permissions", "maintenance_permissions", "deprovision_permissions"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.Contains(t, tfvars, "has_break_glass          = false")
		assert.NotContains(t, tfvars, "break_glass_permissions")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_permissions", "maintenance_permissions", "deprovision_permissions", "break_glass_permissions"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.Contains(t, tfvars, "has_break_glass          = true")
		assert.Contains(t, tfvars, `["iam.roles.get"]`)
	})
}

func TestRenderPredefinedRoles(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_predefined_role", "maintenance_predefined_role", "deprovision_predefined_role"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.NotContains(t, tfvars, "break_glass_predefined_role")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_predefined_role", "maintenance_predefined_role", "deprovision_predefined_role", "break_glass_predefined_role"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
	})
}

func TestRenderBreakGlassConditional(t *testing.T) {
	t.Run("without break glass omits break glass vars", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		assert.Contains(t, tfvars, "has_break_glass          = false")
		assert.NotContains(t, tfvars, "break_glass_permissions")
		assert.NotContains(t, tfvars, "break_glass_predefined_role")
	})

	t.Run("with break glass includes break glass vars", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		assert.Contains(t, tfvars, "has_break_glass          = true")
		assert.Contains(t, tfvars, "break_glass_permissions")
		assert.Contains(t, tfvars, "break_glass_predefined_role")
	})
}

func TestRenderChecksumDiffers(t *testing.T) {
	_, checksum1, err := Render(testInput())
	require.NoError(t, err)

	_, checksum2, err := Render(testInputWithBreakGlass())
	require.NoError(t, err)

	assert.NotEqual(t, checksum1, checksum2, "different inputs should produce different checksums")
}

func TestRenderPredefinedRoleValues(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig = app.AppPermissionsConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				CloudPlatform: "gcp",
				Type:          app.AWSIAMRoleTypeRunnerProvision,
				Policies: []app.AppAWSIAMPolicyConfig{
					{GCPPredefinedRole: "roles/editor"},
				},
			},
		},
	}

	out, _, err := Render(inp)
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	// Find the line with provision_predefined_role and verify value.
	for _, line := range strings.Split(tfvars, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "provision_predefined_role") {
			assert.Contains(t, line, `"roles/editor"`)
		}
	}
}
