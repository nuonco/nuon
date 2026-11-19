package plantypes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositePlanFromAny_OpentofuDeployPlan(t *testing.T) {
	sdkPlan := map[string]any{
		"deploy_plan": map[string]any{
			"opentofu": map[string]any{
				"vars":     map[string]any{"region": "us-east-1"},
				"env_vars": map[string]string{"TF_LOG": "info"},
				"opentofu_backend": map[string]any{
					"WorkspaceID": "ws-1",
				},
				"cluster_info": map[string]any{
					"id":       "eks-cluster",
					"endpoint": "https://1.2.3.4",
					"aws_auth": map[string]any{
						"region": "us-west-2",
						"assume_role": map[string]any{
							"role_arn":     "arn:aws:iam::123:role/deploy",
							"session_name": "deploy",
						},
					},
				},
			},
		},
	}

	cp, err := CompositePlanFromAny(sdkPlan)
	require.NoError(t, err)
	require.NotNil(t, cp.DeployPlan)
	require.NotNil(t, cp.DeployPlan.OpentofuDeployPlan)
	assert.Nil(t, cp.DeployPlan.TerraformDeployPlan)
	assert.Nil(t, cp.DeployPlan.PulumiDeployPlan)

	plan := cp.DeployPlan.OpentofuDeployPlan
	assert.Equal(t, "us-east-1", plan.Vars["region"])
	require.NotNil(t, plan.OpentofuBackend)
	assert.Equal(t, "ws-1", plan.OpentofuBackend.WorkspaceID)

	ci := plan.ClusterInfo
	require.NotNil(t, ci)
	require.NotNil(t, ci.AWSAuth)
	require.NotNil(t, ci.AWSAuth.AssumeRole)
	assert.Equal(t, "arn:aws:iam::123:role/deploy", ci.AWSAuth.AssumeRole.RoleARN)
}

func TestCompositePlanFromAny_PrunesEmptyOpentofuClusterInfo(t *testing.T) {
	sdkPlan := map[string]any{
		"deploy_plan": map[string]any{
			"opentofu": map[string]any{
				"cluster_info": map[string]any{"aws_auth": map[string]any{}},
			},
		},
	}

	cp, err := CompositePlanFromAny(sdkPlan)
	require.NoError(t, err)
	require.NotNil(t, cp.DeployPlan)
	require.NotNil(t, cp.DeployPlan.OpentofuDeployPlan)
	assert.Nil(t, cp.DeployPlan.OpentofuDeployPlan.ClusterInfo, "empty cluster_info from the SDK round-trip must be pruned")
}

func TestDeployPlanOpentofuJSONKey(t *testing.T) {
	bytes, err := json.Marshal(&DeployPlan{
		OpentofuDeployPlan: &OpentofuDeployPlan{
			OpentofuBackend: &TerraformBackend{WorkspaceID: "ws-1"},
		},
	})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(bytes, &doc))
	require.Contains(t, doc, "opentofu", "the opentofu deploy plan must use the opentofu JSON key")

	var dp DeployPlan
	require.NoError(t, json.Unmarshal(bytes, &dp))
	require.NotNil(t, dp.OpentofuDeployPlan)
	assert.Nil(t, dp.TerraformDeployPlan)
}
