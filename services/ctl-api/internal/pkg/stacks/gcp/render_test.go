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

func TestRenderValidJSON(t *testing.T) {
	out, checksum, err := Render(testInput())
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &parsed), "rendered template must be valid JSON")
}

func TestRenderCustomRoles(t *testing.T) {
	out, _, err := Render(testInput())
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &parsed))

	resources := parsed["resource"].(map[string]interface{})
	customRoles, ok := resources["google_project_iam_custom_role"].(map[string]interface{})
	require.True(t, ok, "google_project_iam_custom_role must exist")

	for _, role := range []string{"provision", "maintenance", "deprovision", "break_glass"} {
		assert.Contains(t, customRoles, role, "custom role %s must exist", role)
	}
}

func TestRenderServiceAccounts(t *testing.T) {
	out, _, err := Render(testInput())
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &parsed))

	resources := parsed["resource"].(map[string]interface{})
	sas, ok := resources["google_service_account"].(map[string]interface{})
	require.True(t, ok, "google_service_account must exist")

	for _, sa := range []string{"runner", "provision", "maintenance", "deprovision", "break_glass"} {
		assert.Contains(t, sas, sa, "service account %s must exist", sa)
	}
}

func TestRenderTokenCreatorGrants(t *testing.T) {
	out, _, err := Render(testInput())
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &parsed))

	resources := parsed["resource"].(map[string]interface{})
	iamMembers, ok := resources["google_service_account_iam_member"].(map[string]interface{})
	require.True(t, ok, "google_service_account_iam_member must exist")

	for _, grant := range []string{"provision_token_creator", "maintenance_token_creator", "deprovision_token_creator", "break_glass_token_creator"} {
		assert.Contains(t, iamMembers, grant, "token creator grant %s must exist", grant)
	}
}

func TestRenderOutputs(t *testing.T) {
	out, _, err := Render(testInput())
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &parsed))

	outputs, ok := parsed["output"].(map[string]interface{})
	require.True(t, ok, "outputs must exist")

	for _, output := range []string{"provision_sa_email", "maintenance_sa_email", "deprovision_sa_email", "break_glass_sa_email", "runner_service_account_email"} {
		assert.Contains(t, outputs, output, "output %s must exist", output)
	}
}
