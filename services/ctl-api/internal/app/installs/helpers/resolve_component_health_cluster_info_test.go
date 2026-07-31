package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestComponentHealthClusterInfo(t *testing.T) {
	const roleID = "arn:aws:iam::1234:role/chosen"

	t.Run("aws reads through the chosen role", func(t *testing.T) {
		ci := componentHealthClusterInfo(&app.InstallStackOutputs{
			AWSStackOutputs: &app.AWSStackOutputs{
				Region:                "us-west-2",
				MaintenanceIAMRoleARN: "arn:aws:iam::1234:role/maintenance",
			},
		}, roleID)
		require.NotNil(t, ci)
		require.NotNil(t, ci.AWSAuth)
		require.NotNil(t, ci.AWSAuth.AssumeRole)

		assert.Equal(t, roleID, ci.AWSAuth.AssumeRole.RoleARN)
		assert.Equal(t, componentHealthSessionName, ci.AWSAuth.AssumeRole.SessionName)
		assert.Equal(t, "us-west-2", ci.AWSAuth.Region)
	})

	t.Run("azure reads through the chosen managed identity", func(t *testing.T) {
		ci := componentHealthClusterInfo(&app.InstallStackOutputs{
			AzureStackOutputs: &app.AzureStackOutputs{SubscriptionID: "sub-1"},
		}, "client-id-1")
		require.NotNil(t, ci)
		require.NotNil(t, ci.AzureAuth)

		assert.Equal(t, "client-id-1", ci.AzureAuth.ManagedIdentityClientID)
		assert.Equal(t, "sub-1", ci.AzureAuth.ServicePrincipal.SubscriptionID)
	})

	t.Run("gcp impersonates the chosen service account", func(t *testing.T) {
		ci := componentHealthClusterInfo(&app.InstallStackOutputs{
			GCPStackOutputs: &app.GCPStackOutputs{ProjectID: "proj-1"},
		}, "sa@proj.iam.gserviceaccount.com")
		require.NotNil(t, ci)
		require.NotNil(t, ci.GCPAuth)

		assert.Equal(t, "sa@proj.iam.gserviceaccount.com", ci.GCPAuth.ImpersonateServiceAccount)
	})

	t.Run("no recognized cloud yields no cluster", func(t *testing.T) {
		assert.Nil(t, componentHealthClusterInfo(&app.InstallStackOutputs{}, roleID))
	})
}

// The templated fields are worthless if they don't resolve against real state,
// so assert the round trip rather than the template strings.
func TestComponentHealthClusterInfoRenders(t *testing.T) {
	ci := componentHealthClusterInfo(&app.InstallStackOutputs{
		AWSStackOutputs: &app.AWSStackOutputs{Region: "us-west-2"},
	}, "arn:aws:iam::1234:role/chosen")
	require.NotNil(t, ci)

	stateData := map[string]any{
		"sandbox": map[string]any{
			"outputs": map[string]any{
				"cluster": map[string]any{
					"name":                       "my-cluster",
					"endpoint":                   "https://eks.example.com",
					"certificate_authority_data": "Y2FkYXRh",
				},
			},
		},
	}

	require.NoError(t, render.RenderStruct(ci, stateData))
	assert.Equal(t, "my-cluster", ci.ID)
	assert.Equal(t, "https://eks.example.com", ci.Endpoint)
	assert.Equal(t, "Y2FkYXRh", ci.CAData)
}

func TestSandboxEmitsClusterOutputs(t *testing.T) {
	cases := map[string]struct {
		state map[string]any
		want  bool
	}{
		"cluster present": {
			state: map[string]any{"sandbox": map[string]any{"outputs": map[string]any{"cluster": map[string]any{"name": "c"}}}},
			want:  true,
		},
		"cluster nil": {
			state: map[string]any{"sandbox": map[string]any{"outputs": map[string]any{"cluster": nil}}},
			want:  false,
		},
		"no cluster key": {
			state: map[string]any{"sandbox": map[string]any{"outputs": map[string]any{"vpc": "v"}}},
			want:  false,
		},
		"no sandbox": {
			state: map[string]any{},
			want:  false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, sandboxEmitsClusterOutputs(tc.state))
		})
	}
}
