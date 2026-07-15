package plantypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestCompositePlanFromAny_PrunesEmptyAWSAuth(t *testing.T) {
	// Mirrors a GCP install plan after passing through the runner SDK:
	// models.KubeClusterInfo.AwsAuth is a struct value, so a null aws_auth
	// from the API re-marshals as {}.
	sdkPlan := &models.PlantypesCompositePlan{
		DeployPlan: &models.PlantypesDeployPlan{
			Terraform: &models.PlantypesTerraformDeployPlan{
				ClusterInfo: &models.KubeClusterInfo{
					ID:       "gke-cluster",
					Endpoint: "https://1.2.3.4",
					GcpAuth: &models.GithubComNuoncoNuonPkgGcpCredentialsConfig{
						ProjectID: "proj",
						Region:    "us-central1",
					},
				},
			},
		},
	}

	cp, err := CompositePlanFromAny(sdkPlan)
	require.NoError(t, err)
	require.NotNil(t, cp.DeployPlan)
	require.NotNil(t, cp.DeployPlan.TerraformDeployPlan)

	ci := cp.DeployPlan.TerraformDeployPlan.ClusterInfo
	require.NotNil(t, ci)
	assert.Nil(t, ci.AWSAuth, "empty aws_auth from the SDK round-trip must be pruned")
	require.NotNil(t, ci.GCPAuth)
	assert.Equal(t, "proj", ci.GCPAuth.ProjectID)
}

func TestCompositePlanFromAny_PrunesEmptyClusterInfo(t *testing.T) {
	// Mirrors an action workflow run plan for a sandbox with no cluster:
	// models.PlantypesActionWorkflowRunPlan.ClusterInfo is a struct value, so a
	// null cluster_info from the API re-marshals as {} (with an empty aws_auth
	// under the old SDK models).
	sdkPlan := map[string]any{
		"action_workflow_run_plan": map[string]any{
			"id":           "run-1",
			"install_id":   "inst-1",
			"cluster_info": map[string]any{"aws_auth": map[string]any{}},
		},
	}

	cp, err := CompositePlanFromAny(sdkPlan)
	require.NoError(t, err)
	require.NotNil(t, cp.ActionWorkflowRunPlan)
	assert.Nil(t, cp.ActionWorkflowRunPlan.ClusterInfo, "empty cluster_info from the SDK round-trip must be pruned")
}

func TestCompositePlanFromAny_KeepsPopulatedAWSAuth(t *testing.T) {
	sdkPlan := map[string]any{
		"deploy_plan": map[string]any{
			"terraform": map[string]any{
				"cluster_info": map[string]any{
					"id": "eks-cluster",
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

	ci := cp.DeployPlan.TerraformDeployPlan.ClusterInfo
	require.NotNil(t, ci)
	require.NotNil(t, ci.AWSAuth)
	require.NotNil(t, ci.AWSAuth.AssumeRole)
	assert.Equal(t, "arn:aws:iam::123:role/deploy", ci.AWSAuth.AssumeRole.RoleARN)
}

func TestCompositePlanFromAny_KeepsHelmSkipCRDs(t *testing.T) {
	sdkPlan := &models.PlantypesCompositePlan{
		DeployPlan: &models.PlantypesDeployPlan{
			Helm: &models.PlantypesHelmDeployPlan{
				SkipCrds: true,
			},
		},
	}

	cp, err := CompositePlanFromAny(sdkPlan)
	require.NoError(t, err)
	require.NotNil(t, cp.DeployPlan)
	require.NotNil(t, cp.DeployPlan.HelmDeployPlan)
	assert.True(t, cp.DeployPlan.HelmDeployPlan.SkipCRDs)
}
