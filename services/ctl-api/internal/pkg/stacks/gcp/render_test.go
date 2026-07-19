package gcp

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/config"
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

// extractTfvars parses the JSON envelope and returns the inputs tfvars string
// (standard vars, permissions, roles, install_inputs).
func extractTfvars(t *testing.T, out []byte) string {
	t.Helper()
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(out, &envelope))
	tfvars, ok := envelope["inputs_tfvars"]
	require.True(t, ok, "envelope must contain 'inputs_tfvars' key")
	return tfvars
}

// extractSecretsTfvars parses the JSON envelope and returns the secrets tfvars
// string (auto_generate_secrets, secrets).
func extractSecretsTfvars(t *testing.T, out []byte) string {
	t.Helper()
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(out, &envelope))
	tfvars, ok := envelope["secrets_tfvars"]
	require.True(t, ok, "envelope must contain 'secrets_tfvars' key")
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

func TestRenderGCPAccountInjection(t *testing.T) {
	t.Run("with project_id and region", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.Contains(t, tfvars, `gcp_project_id           = "my-gcp-project"`)
		assert.Contains(t, tfvars, `gcp_region               = "us-central1"`)
	})

	t.Run("with empty project_id and region", func(t *testing.T) {
		inp := testInput()
		inp.Install.GCPAccount = &app.GCPAccount{}
		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.NotContains(t, tfvars, "gcp_project_id")
		assert.NotContains(t, tfvars, "gcp_region")
	})

	t.Run("with nil GCPAccount", func(t *testing.T) {
		inp := testInput()
		inp.Install.GCPAccount = nil
		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.NotContains(t, tfvars, "gcp_project_id")
		assert.NotContains(t, tfvars, "gcp_region")
	})
}

func TestRenderMachineType(t *testing.T) {
	t.Run("with instance type set", func(t *testing.T) {
		inp := testInput()
		inp.Settings.AWSInstanceType = "e2-medium"
		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.Contains(t, tfvars, `runner_machine_type      = "e2-medium"`)
	})

	t.Run("without instance type", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.NotContains(t, tfvars, "runner_machine_type")
	})
}

func TestRenderPermissions(t *testing.T) {
	t.Run("without break glass", func(t *testing.T) {
		out, _, err := Render(testInput())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_policies", "maintenance_policies", "deprovision_policies"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.Contains(t, tfvars, "break_glass_roles = {\n}")
	})

	t.Run("with break glass", func(t *testing.T) {
		out, _, err := Render(testInputWithBreakGlass())
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)

		for _, v := range []string{"provision_policies", "maintenance_policies", "deprovision_policies"} {
			assert.Contains(t, tfvars, v+" ", "tfvars should contain %s", v)
		}
		assert.Contains(t, tfvars, `"emergency-access"`)
		assert.Contains(t, tfvars, `["iam.roles.get"]`)
		assert.Contains(t, tfvars, "enabled         = false")
	})

	t.Run("standard role policies stay separate", func(t *testing.T) {
		inp := testInput()
		inp.AppCfg.PermissionsConfig = app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "gcp",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{Name: "provision-network", GCPPermissions: []string{"compute.networks.create"}},
						{Name: "provision-dns", GCPPermissions: []string{"dns.changes.create"}},
						{Name: "provision-gke", GCPPredefinedRole: "roles/container.admin"},
					},
				},
			},
		}

		out, _, err := Render(inp)
		require.NoError(t, err)

		tfvars := extractTfvars(t, out)
		assert.Contains(t, tfvars, `"provision-network" = ["compute.networks.create"]`)
		assert.Contains(t, tfvars, `"provision-dns" = ["dns.changes.create"]`)
		assert.Contains(t, tfvars, `provision_predefined_role    = "roles/container.admin"`)
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

		tfvars := extractSecretsTfvars(t, out)

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

		tfvars := extractSecretsTfvars(t, out)

		assert.Contains(t, tfvars, "auto_generate_secrets = []")
		assert.Contains(t, tfvars, "secrets = {\n}")
	})
}

func extractEnvelopeKey(t *testing.T, out []byte, key string) string {
	t.Helper()
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(out, &envelope))
	val, ok := envelope[key]
	require.True(t, ok, "envelope must contain %q key", key)
	return val
}

func TestRenderSpaceliftArtifacts(t *testing.T) {
	inp := testInput()
	inp.AppCfg.InputConfig = app.AppInputConfig{
		AppInputs: []app.AppInput{
			{Name: "cluster_name", Source: app.AppInputSourceCustomer},
		},
	}
	inp.AppCfg.SecretsConfig = app.AppSecretsConfig{
		Secrets: []app.AppSecretConfig{
			{Name: "stripe_key", Description: "Your Stripe API key", Required: true},
		},
	}

	out, _, err := Render(inp)
	require.NoError(t, err)

	adminTF := extractEnvelopeKey(t, out, "spacelift_admin_tf")
	blueprint := extractEnvelopeKey(t, out, "spacelift_blueprint_yaml")

	require.NotEmpty(t, adminTF)
	require.NotEmpty(t, blueprint)

	for _, artifact := range []string{adminTF, blueprint} {
		assert.Contains(t, artifact, inp.Install.ID)
		assert.Contains(t, artifact, "install-stacks")
	}

	// The admin stack reads the tfvars from sibling files so the customer can
	// edit inputs and replace secrets before applying.
	assert.Contains(t, adminTF, "spacelift_stack")
	assert.Contains(t, adminTF, `project_root`)
	assert.Contains(t, adminTF, `"gcp"`)
	assert.Contains(t, adminTF, "raw_git {")
	assert.Contains(t, adminTF, `url       = "https://github.com/nuonco/install-stacks.git"`)
	assert.Contains(t, adminTF, `relative_path = "source/gcp/inputs.auto.tfvars"`)
	assert.Contains(t, adminTF, `relative_path = "source/gcp/secrets.auto.tfvars"`)
	assert.Contains(t, adminTF, `write_only    = false`, "inputs mounted file should be plain")
	assert.Contains(t, adminTF, `write_only    = true`, "secrets mounted file should be secret")
	assert.Contains(t, adminTF, `filebase64("${path.module}/inputs.auto.tfvars")`)
	assert.Contains(t, adminTF, `filebase64("${path.module}/secrets.auto.tfvars")`)
	assert.Contains(t, adminTF, `variable "space_id"`, "admin stack should require an explicit space_id")
	assert.Contains(t, adminTF, `space_id          = var.space_id`)
	assert.Contains(t, adminTF, `variable "attach_gcp_service_account"`, "GCP integration attach should be toggleable for customers with their own integration")
	assert.Contains(t, adminTF, `resource "spacelift_gcp_service_account" "nuon"`, "admin stack should auto-attach the GCP cloud integration by default")
	assert.Contains(t, adminTF, `count        = var.attach_gcp_service_account ? 1 : 0`)

	assert.Contains(t, blueprint, "project_root: gcp")
	assert.Contains(t, blueprint, "provider: RAW_GIT")
	assert.Contains(t, blueprint, "repository_url: https://github.com/nuonco/install-stacks.git")
	assert.Contains(t, blueprint, "vendor:")
	assert.NotContains(t, blueprint, "trigger_run", "blueprint must not auto-trigger: GCP creds aren't attachable via blueprint, so the first run would fail auth")

	// GCP project/region, customer install inputs, and secrets are exposed as
	// blueprint inputs and interpolated into the (plaintext) mounted tfvars via CEL.
	assert.Contains(t, blueprint, "inputs:")
	assert.Contains(t, blueprint, "id: gcp_project_id")
	assert.Contains(t, blueprint, `default: "my-gcp-project"`)
	assert.Contains(t, blueprint, "id: gcp_region")
	assert.Contains(t, blueprint, `default: "us-central1"`)
	assert.Contains(t, blueprint, "id: input_cluster_name")
	assert.Contains(t, blueprint, "type: short_text")
	assert.Contains(t, blueprint, "id: secret_stripe_key")
	assert.Contains(t, blueprint, "type: secret")
	assert.Contains(t, blueprint, `description: "Your Stripe API key"`)

	assert.Contains(t, blueprint, `gcp_project_id           = "${{ inputs.gcp_project_id }}"`)
	assert.Contains(t, blueprint, `gcp_region               = "${{ inputs.gcp_region }}"`)
	assert.Contains(t, blueprint, `"cluster_name" = "${{ inputs.input_cluster_name }}"`)
	assert.Contains(t, blueprint, `value       = "${{ inputs.secret_stripe_key }}"`)

	// Content is plaintext (not base64) so CEL interpolation works.
	assert.Contains(t, blueprint, "nuon_install_id")
	assert.NotContains(t, blueprint, base64.StdEncoding.EncodeToString([]byte("nuon_install_id")))
}

func TestRenderSpaceliftDefaults(t *testing.T) {
	inp := testInput()
	inp.AppCfg.InputConfig = app.AppInputConfig{
		AppInputs: []app.AppInput{
			{Name: "cluster_name", Source: app.AppInputSourceCustomer, Default: "my-default-cluster"},
		},
	}
	inp.AppCfg.SecretsConfig = app.AppSecretsConfig{
		Secrets: []app.AppSecretConfig{
			{Name: "stripe_key", Description: "Your Stripe API key", Required: true, Default: "sk_default"},
		},
	}

	out, _, err := Render(inp)
	require.NoError(t, err)

	inputsTfvars := extractEnvelopeKey(t, out, "inputs_tfvars")
	secretsTfvars := extractEnvelopeKey(t, out, "secrets_tfvars")
	blueprint := extractEnvelopeKey(t, out, "spacelift_blueprint_yaml")

	// The normal (admin-TF sibling) tfvars pre-fill install-input and secret
	// defaults as literals.
	assert.Contains(t, inputsTfvars, `"cluster_name" = "my-default-cluster"`)
	assert.Contains(t, secretsTfvars, `value       = "sk_default"`)

	// The blueprint surfaces those same defaults as the blueprint input `default`,
	// so the Spacelift UI is pre-populated.
	assert.Contains(t, blueprint, `default: "my-default-cluster"`)
	assert.Contains(t, blueprint, `default: "sk_default"`)

	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(blueprint), &parsed))
}

func TestRenderSpaceliftBlueprintYAMLValid(t *testing.T) {
	inp := testInput()
	inp.AppCfg.InputConfig = app.AppInputConfig{
		AppInputs: []app.AppInput{
			{Name: "cluster_name", Source: app.AppInputSourceCustomer},
		},
	}
	inp.AppCfg.SecretsConfig = app.AppSecretsConfig{
		Secrets: []app.AppSecretConfig{
			// Description contains ": " which, if interpolated unquoted, makes YAML
			// read it as a nested mapping ("mapping values are not allowed in this
			// context"). This mirrors the byoc slack secret descriptions.
			{Name: "slack_signing_secret", Description: "Signing Secret. Managed out-of-band: the central AWS Secrets Manager entry is the source of truth.", Required: false},
		},
	}

	out, _, err := Render(inp)
	require.NoError(t, err)

	blueprint := extractEnvelopeKey(t, out, "spacelift_blueprint_yaml")

	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(blueprint), &parsed),
		"rendered blueprint must be valid YAML even when input descriptions contain ': '")
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

func TestRenderCustomStacks(t *testing.T) {
	inp := testInput()
	inp.AppCfg.StackConfig.CustomNestedStacks = []config.CustomNestedStack{
		{Name: "preview_bucket", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket", Index: 0},
	}

	out, _, err := Render(inp)
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)
	assert.Contains(t, tfvars, "custom_stacks = {")
	assert.Contains(t, tfvars, `"preview_bucket" = {`)
	assert.Contains(t, tfvars, `module = "bucket"`)

	inp.AppCfg.StackConfig.CustomNestedStacks[0].TemplateURL = "https://example.com/stack.yaml"
	_, _, err = Render(inp)
	require.Error(t, err)
}
