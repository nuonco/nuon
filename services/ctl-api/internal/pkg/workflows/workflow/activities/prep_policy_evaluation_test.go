package activities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPreparePolicyInputsPulumi(t *testing.T) {
	preview := []byte(`{
		"stdout": "preview",
		"change_summary": {"create": 1},
		"resource_changes": [{
			"urn": "urn:pulumi:dev::app::aws:s3/bucket:Bucket::assets",
			"type": "aws:s3/bucket:Bucket",
			"name": "assets",
			"action": "create",
			"new_inputs": {"acl": "private"}
		}]
	}`)

	tests := map[string]*policyContext{
		"component": {
			ComponentType: app.ComponentTypePulumi,
			ComponentID:   ptrString("component-id"),
			ComponentName: "infrastructure",
		},
		"sandbox": {
			IsSandbox:   true,
			SandboxType: config.AppSandboxTypePulumi,
		},
	}

	for name, pctx := range tests {
		t.Run(name, func(t *testing.T) {
			inputs, identities, err := (&Activities{}).preparePolicyInputs(preview, pctx)
			require.NoError(t, err)
			require.Len(t, inputs, 1)
			require.Len(t, identities, 1)

			var input map[string]any
			require.NoError(t, json.Unmarshal(inputs[0], &input))
			assert.Contains(t, input, "resource_changes")
			assert.NotContains(t, input, "plan")
		})
	}
}

func TestPreparePolicyInputsTerraformSandbox(t *testing.T) {
	inputs, _, err := (&Activities{}).preparePolicyInputs([]byte(fakeTerraformPlanDisplayContents), &policyContext{
		IsSandbox:   true,
		SandboxType: config.AppSandboxTypeTerraform,
	})
	require.NoError(t, err)
	require.Len(t, inputs, 1)

	var input map[string]any
	require.NoError(t, json.Unmarshal(inputs[0], &input))
	assert.Contains(t, input, "plan")
}

func TestPreparePulumiPolicyInputsRejectsWrongShape(t *testing.T) {
	_, _, err := (&Activities{}).preparePulumiPolicyInputs([]byte(`{"resource_changes": []}`), &policyContext{})
	require.ErrorContains(t, err, "missing change_summary")
}

func TestPrepareKubernetesPolicyInputsRejectEmptyRender(t *testing.T) {
	tests := map[string]func() error{
		"helm": func() error {
			_, _, err := (&Activities{}).prepareHelmPolicyInputs([]byte(`{"op":"install","template_output":""}`))
			return err
		},
		"kubernetes manifest": func() error {
			_, _, err := (&Activities{}).prepareKubernetesManifestPolicyInputs([]byte(`{"op":"apply","dry_run_output":""}`))
			return err
		},
	}

	for name, evaluate := range tests {
		t.Run(name, func(t *testing.T) {
			require.ErrorContains(t, evaluate(), "contains no Kubernetes manifests")
		})
	}
}

func TestPrepareKubernetesManifestDeletePolicyInputs(t *testing.T) {
	planContents := []byte(`{
		"op": "delete",
		"dry_run_output": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: settings\n  namespace: default\n"
	}`)

	inputs, identities, err := (&Activities{}).prepareKubernetesManifestPolicyInputs(planContents)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, []string{"ConfigMap/default/settings"}, identities)

	var input map[string]any
	require.NoError(t, json.Unmarshal(inputs[0], &input))
	review := input["review"].(map[string]any)
	assert.Equal(t, "DELETE", review["operation"])
	assert.Equal(t, "settings", review["object"].(map[string]any)["metadata"].(map[string]any)["name"])
}

func ptrString(value string) *string {
	return &value
}
