package aws

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
			AWSAccount: &app.AWSAccount{
				Region: "us-east-1",
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

// extractTfvars unwraps the JSON envelope into the raw HCL tfvars string.
func extractTfvars(t *testing.T, out []byte) string {
	t.Helper()
	var envelope map[string]string
	require.NoError(t, json.Unmarshal(out, &envelope))
	return envelope["tfvars"]
}

// findVarValue finds the right-hand side of `<key> = ...` in the HCL tfvars.
// Returns the trimmed value (possibly a quoted JSON string).
func findVarValue(t *testing.T, tfvars, key string) string {
	t.Helper()
	for _, line := range strings.Split(tfvars, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, key) {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq == -1 {
			continue
		}
		return strings.TrimSpace(trimmed[eq+1:])
	}
	t.Fatalf("key %q not found in tfvars:\n%s", key, tfvars)
	return ""
}

// unquoteHCLJSONString takes an HCL string literal that wraps a JSON document
// (the form mergedInlinePolicyDocument emits) and returns the inner JSON.
func unquoteHCLJSONString(t *testing.T, s string) string {
	t.Helper()
	var inner string
	require.NoError(t, json.Unmarshal([]byte(s), &inner), "value must be a JSON-quoted string: %s", s)
	return inner
}

func TestRenderValidJSON(t *testing.T) {
	out, checksum, err := Render(testInput(), "")
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(out, &parsed), "rendered envelope must be valid JSON")
	_, ok := parsed["tfvars"].(string)
	assert.True(t, ok, "envelope must contain a tfvars string")
}

func TestRenderStandardVars(t *testing.T) {
	out, _, err := Render(testInput(), "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	for key, val := range map[string]string{
		"nuon_install_id":        `"instabcdefghijklmnopqrstuv"`,
		"nuon_org_id":            `"orgabcdefghijklmnopqrstuvw"`,
		"nuon_app_id":            `"appabcdefghijklmnopqrstuvw"`,
		"runner_api_url":         `"https://runner.nuon.co"`,
		"runner_api_token":       `"test-token"`,
		"runner_id":              `"runnerabcdefghijklmnopqrstu"`,
		"runner_init_script_url": `"https://example.com/init.sh"`,
		"phone_home_url":         `"https://example.com/phone-home"`,
		"aws_region":             `"us-east-1"`,
	} {
		assert.Equal(t, val, findVarValue(t, tfvars, key), "key %s", key)
	}
}

func TestRenderEmptyPermissions(t *testing.T) {
	out, _, err := Render(testInput(), "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	assert.Equal(t, "[]", findVarValue(t, tfvars, "provision_permissions"))
	assert.Equal(t, "[]", findVarValue(t, tfvars, "maintenance_permissions"))
	assert.Equal(t, "[]", findVarValue(t, tfvars, "deprovision_permissions"))
	assert.Equal(t, `""`, findVarValue(t, tfvars, "provision_inline_policy_document"))
	assert.Equal(t, `""`, findVarValue(t, tfvars, "maintenance_inline_policy_document"))
	assert.Equal(t, `""`, findVarValue(t, tfvars, "deprovision_inline_policy_document"))
	assert.Equal(t, "[]", findVarValue(t, tfvars, "provision_managed_policy_arns"))
	assert.Equal(t, "[]", findVarValue(t, tfvars, "maintenance_managed_policy_arns"))
	assert.Equal(t, "[]", findVarValue(t, tfvars, "deprovision_managed_policy_arns"))
}

func TestRenderManagedPolicyOnlyRole(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{ManagedPolicyName: "AdministratorAccess"},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	assert.Equal(t,
		`["arn:aws:iam::aws:policy/AdministratorAccess"]`,
		findVarValue(t, tfvars, "provision_managed_policy_arns"),
	)
	assert.Equal(t, `""`, findVarValue(t, tfvars, "provision_inline_policy_document"))
}

func TestRenderInlinePolicyOnlyRole(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{
					Name: "scoped-s3",
					Contents: []byte(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Effect": "Allow",
							"Action": ["s3:PutObject", "s3:GetObject"],
							"Resource": "arn:aws:s3:::vendor-bucket/*",
							"Condition": {"StringEquals": {"s3:x-amz-acl": "bucket-owner-full-control"}}
						}]
					}`),
				},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	// Managed-arns slot stays empty.
	assert.Equal(t, "[]", findVarValue(t, tfvars, "provision_managed_policy_arns"))

	// Inline document round-trips with full fidelity.
	doc := unquoteHCLJSONString(t, findVarValue(t, tfvars, "provision_inline_policy_document"))
	var parsed struct {
		Version   string
		Statement []map[string]any
	}
	require.NoError(t, json.Unmarshal([]byte(doc), &parsed))
	assert.Equal(t, "2012-10-17", parsed.Version)
	require.Len(t, parsed.Statement, 1)
	assert.Equal(t, "Allow", parsed.Statement[0]["Effect"])
	assert.Equal(t, "arn:aws:s3:::vendor-bucket/*", parsed.Statement[0]["Resource"])
	assert.NotNil(t, parsed.Statement[0]["Condition"])
}

func TestRenderMixedManagedAndInline(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{ManagedPolicyName: "ReadOnlyAccess"},
				{
					Name:     "extra",
					Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"ec2:RunInstances","Resource":"*"}]}`),
				},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)

	assert.Equal(t,
		`["arn:aws:iam::aws:policy/ReadOnlyAccess"]`,
		findVarValue(t, tfvars, "provision_managed_policy_arns"),
	)
	doc := unquoteHCLJSONString(t, findVarValue(t, tfvars, "provision_inline_policy_document"))
	assert.Contains(t, doc, `"ec2:RunInstances"`)
}

func TestRenderInlinePolicyMergesAcrossPolicies(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{Name: "a", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`)},
				{Name: "b", Contents: []byte(`{"Statement":[{"Effect":"Deny","Action":"s3:DeleteObject","Resource":"*"}]}`)},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	doc := unquoteHCLJSONString(t, findVarValue(t, extractTfvars(t, out), "provision_inline_policy_document"))
	var parsed struct {
		Statement []map[string]any
	}
	require.NoError(t, json.Unmarshal([]byte(doc), &parsed))
	require.Len(t, parsed.Statement, 2, "statements from multiple policies should be merged")
	assert.Equal(t, "Allow", parsed.Statement[0]["Effect"])
	assert.Equal(t, "Deny", parsed.Statement[1]["Effect"])
}

func TestRenderInlinePolicyActionAsStringOrSlice(t *testing.T) {
	// Both forms are valid IAM policy JSON. Renderer must round-trip both
	// without mangling — Action stays whatever shape AWS gave us.
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{Name: "single", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`)},
				{Name: "multi", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":["sqs:SendMessage","sqs:ReceiveMessage"],"Resource":"*"}]}`)},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	doc := unquoteHCLJSONString(t, findVarValue(t, extractTfvars(t, out), "provision_inline_policy_document"))
	assert.Contains(t, doc, `"Action":"s3:GetObject"`)
	assert.Contains(t, doc, `"Action":["sqs:SendMessage","sqs:ReceiveMessage"]`)
}

func TestRenderInlinePolicyMalformedReturnsError(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{Name: "bad", Contents: []byte(`not json at all`)},
			},
		},
	}

	_, _, err := Render(inp, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provision")
}

func TestRenderBreakGlassInlinePolicy(t *testing.T) {
	inp := testInput()
	inp.AppCfg.BreakGlassConfig = app.AppBreakGlassConfig{
		Roles: []app.AppAWSIAMRoleConfig{
			{
				CloudPlatform: "aws",
				Type:          app.AWSIAMRoleTypeBreakGlass,
				Name:          "emergency",
				Policies: []app.AppAWSIAMPolicyConfig{
					{Name: "p", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"iam:*","Resource":"*"}]}`)},
				},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)
	assert.Contains(t, tfvars, `"emergency"`)
	// inline_policy_document is an HCL-quoted JSON string; look at the unwrapped doc.
	doc := unquoteHCLJSONString(t, findVarValue(t, tfvars, "inline_policy_document"))
	assert.Contains(t, doc, `"iam:*"`)
	// Break-glass roles default to disabled.
	assert.Contains(t, tfvars, "enabled                = false")
}

func TestRenderCustomRoleInlinePolicy(t *testing.T) {
	inp := testInput()
	inp.AppCfg.PermissionsConfig.CustomRoles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeCustom,
			Name:          "db-reader",
			Policies: []app.AppAWSIAMPolicyConfig{
				{Name: "p", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"rds:Describe*","Resource":"*"}]}`)},
			},
		},
	}

	out, _, err := Render(inp, "")
	require.NoError(t, err)

	tfvars := extractTfvars(t, out)
	assert.Contains(t, tfvars, `"db-reader"`)
	doc := unquoteHCLJSONString(t, findVarValue(t, tfvars, "inline_policy_document"))
	assert.Contains(t, doc, `"rds:Describe*"`)
	assert.Contains(t, tfvars, "enabled                = true")
}

func TestRenderChecksumDiffersWithInlinePolicy(t *testing.T) {
	_, base, err := Render(testInput(), "")
	require.NoError(t, err)

	inp := testInput()
	inp.AppCfg.PermissionsConfig.Roles = []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "aws",
			Type:          app.AWSIAMRoleTypeRunnerProvision,
			Name:          "provision",
			Policies: []app.AppAWSIAMPolicyConfig{
				{Name: "p", Contents: []byte(`{"Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`)},
			},
		},
	}
	_, withInline, err := Render(inp, "")
	require.NoError(t, err)

	assert.NotEqual(t, base, withInline, "inline policy should affect the checksum")
}
