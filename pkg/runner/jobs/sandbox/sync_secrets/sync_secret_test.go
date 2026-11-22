package terraform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/outputs"
)

// TestExecSyncSecretV2_Expansion verifies the v2 path fans a single source out across every target × namespace,
// writing one output per destination keyed uniquely so destinations sharing a name (across namespaces, or same
// namespace/name with different keys) do not overwrite one another. exists=false keeps the Kubernetes upsert out of
// the test (no cluster needed); only the expansion + output keying is exercised.
func TestExecSyncSecretV2_Expansion(t *testing.T) {
	h := &handler{state: &handlerState{outputs: make(outputs.SyncSecretsOutput)}}

	secr := plantypes.KubernetesSecretSync{
		SecretName: "datadog-api-key",
		SecretARN:  "arn:aws:secretsmanager:us-west-2:123:secret:dd",
		Targets: []plantypes.KubernetesSecretSyncTarget{
			{Namespaces: []string{"cloudprem", "datadog"}, Name: "datadog", Key: "api-key"},
			{Namespaces: []string{"datadog"}, Name: "datadog", Key: "app-key"},
		},
	}

	err := h.execSyncSecretV2(context.Background(), secr, "", nil, false)
	require.NoError(t, err)

	// 2 namespaces on target 1 + 1 namespace on target 2 = 3 unique destinations.
	require.Len(t, h.state.outputs, 3)

	cloudpremAPI, ok := h.state.outputs["datadog-api-key/cloudprem/datadog/api-key"]
	require.True(t, ok)
	assert.Equal(t, "cloudprem", cloudpremAPI.KubernetesNamespace)
	assert.Equal(t, "datadog", cloudpremAPI.KubernetesName)
	assert.Equal(t, "api-key", cloudpremAPI.KubernetesKey)
	assert.Equal(t, "datadog-api-key", cloudpremAPI.Name)
	assert.Equal(t, "arn:aws:secretsmanager:us-west-2:123:secret:dd", cloudpremAPI.ARN)
	assert.False(t, cloudpremAPI.Exists)

	_, ok = h.state.outputs["datadog-api-key/datadog/datadog/api-key"]
	assert.True(t, ok)

	// same namespace/name as the api-key destination but a different key — must be its own entry, not a clobber.
	_, ok = h.state.outputs["datadog-api-key/datadog/datadog/app-key"]
	assert.True(t, ok)
}

// TestExecSyncSecretV1_SingleOutput verifies the legacy single-destination path records exactly one output keyed by
// the v1 destination.
func TestExecSyncSecretV1_SingleOutput(t *testing.T) {
	h := &handler{state: &handlerState{outputs: make(outputs.SyncSecretsOutput)}}

	secr := plantypes.KubernetesSecretSync{
		SecretName: "db",
		SecretARN:  "arn:aws:secretsmanager:us-west-2:123:secret:db",
		Namespace:  "default",
		Name:       "app-db",
		KeyName:    "value",
	}

	err := h.execSyncSecretV1(context.Background(), secr, "", nil, false)
	require.NoError(t, err)

	require.Len(t, h.state.outputs, 1)
	out, ok := h.state.outputs["db/default/app-db/value"]
	require.True(t, ok)
	assert.Equal(t, "default", out.KubernetesNamespace)
	assert.Equal(t, "app-db", out.KubernetesName)
	assert.Equal(t, "value", out.KubernetesKey)
}
