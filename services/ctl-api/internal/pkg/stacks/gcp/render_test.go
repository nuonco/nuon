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
				Name:          "emergency-access",
				Policies: []app.AppAWSIAMPolicyConfig{
					{GCPPermissions: []string{"iam.roles.get"}},
				},
			},
		},
	}
	return inp
}

func testInputWithMultipleBreakGlass() *stacks.TemplateInput {
	inp := testInput()
	inp.AppCfg.BreakGlassConfig = app.AppBreakGlassConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				CloudPlatform: "gcp",
				Type:          app.AWSIAMRoleTypeBreakGlass,
				Name:          "emergency-access",
				Policies: []app.AppAWSIAMPolicyConfig{
					{GCPPermissions: []string{"iam.roles.get"}},
				},
			},
			{
				CloudPlatform: "gcp",
				Type:          app.AWSIAMRoleTypeBreakGlass,
				Name:          "admin-access",
				Policies: []app.AppAWSIAMPolicyConfig{
					{GCPPermissions: []string{"compute.instances.list", "storage.buckets.list"}},
				},
			},
		},
	}
	return inp
}

func testInputWithCustomRoles() *stacks.TemplateInput {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.CustomRoles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "gcp",
			Type:          app.AWSIAMRoleTypeCustom,
			Name:          "db-reader",
			Policies: []app.AppAWSIAMPolicyConfig{
				{GCPPermissions: []string{"cloudsql.instances.list"}},
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
		assert.Contains(t, tfvars, "break_glass_roles = {\n}")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_permissions", "maintenance_permissions", "deprovision_permissions"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.Contains(t, tfvars, `"emergency-access"`)
		assert.Contains(t, tfvars, `["iam.roles.get"]`)
		assert.Contains(t, tfvars, "enabled         = false")
	})
}

func TestRenderMultipleBreakGlassRoles(t *testing.T) {
	out, _, err := Render(testInputWithMultipleBreakGlass())
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	assert.Contains(t, tfvars, `"emergency-access"`)
	assert.Contains(t, tfvars, `"admin-access"`)
	assert.Contains(t, tfvars, `["iam.roles.get"]`)
	assert.Contains(t, tfvars, `["compute.instances.list","storage.buckets.list"]`)

	// Both should be disabled by default
	count := strings.Count(tfvars, "enabled         = false")
	assert.Equal(t, 2, count, "both breakglass roles should be disabled by default")
}

func TestRenderCustomRoles(t *testing.T) {
	out, _, err := Render(testInputWithCustomRoles())
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	assert.Contains(t, tfvars, `"db-reader"`)
	assert.Contains(t, tfvars, `["cloudsql.instances.list"]`)
	assert.Contains(t, tfvars, "enabled         = true")
}

func TestRenderPredefinedRoles(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_predefined_role", "maintenance_predefined_role", "deprovision_predefined_role"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
	})

	t.Run("with break glass predefined role", func(t *testing.T) {
		inp := testInput()
		inp.AppCfg.BreakGlassConfig = app.AppBreakGlassConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "gcp",
					Type:          app.AWSIAMRoleTypeBreakGlass,
					Name:          "elevated-access",
					Policies: []app.AppAWSIAMPolicyConfig{
						{GCPPredefinedRole: "roles/editor"},
					},
				},
			},
		}

		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		assert.Contains(t, tfvars, `"elevated-access"`)
		assert.Contains(t, tfvars, `predefined_role = "roles/editor"`)
	})
}

func TestRenderChecksumDiffers(t *testing.T) {
	_, checksum1, err := Render(testInput())
	require.NoError(t, err)

	_, checksum2, err := Render(testInputWithBreakGlass())
	require.NoError(t, err)

	assert.NotEqual(t, checksum1, checksum2, "different inputs should produce different checksums")
}

func TestRenderSecrets(t *testing.T) {
	t.Run("auto-gen and customer secrets", func(t *testing.T) {
		inp := testInput()
		inp.AppCfg.SecretsConfig = app.AppSecretsConfig{
			Secrets: []app.AppSecretConfig{
				{
					Name:         "db_password",
					AutoGenerate: true,
				},
				{
					Name:        "stripe_key",
					Description: "Your Stripe API key",
					Required:    true,
				},
				{
					Name:        "optional_key",
					Description: "Optional config",
					Default:     "default-val",
				},
			},
		}

		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		// auto-gen should be in the list
		assert.Contains(t, tfvars, `auto_generate_secrets = ["db_password", ]`)

		// customer secrets should be in the secrets block
		assert.Contains(t, tfvars, `"stripe_key"`)
		assert.Contains(t, tfvars, `description = "Your Stripe API key"`)
		assert.Contains(t, tfvars, `required    = true`)

		assert.Contains(t, tfvars, `"optional_key"`)
		assert.Contains(t, tfvars, `value       = "default-val"`)

		// auto-gen should NOT appear in customer secrets
		assert.NotContains(t, tfvars, `"db_password" = {`)
	})

	t.Run("no secrets", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		assert.Contains(t, tfvars, "auto_generate_secrets = []")
		assert.Contains(t, tfvars, "secrets = {\n}")
	})
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
