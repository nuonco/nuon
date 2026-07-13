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
