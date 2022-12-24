package sandbox

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	shared "github.com/powertoolsdev/workers-installs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestWorkflow_Provision(t *testing.T) {
	cfg := generics.GetFakeObj[shared.Config]()
	assert.Nil(t, cfg.Validate())
	req := generics.GetFakeObj[*sandboxv1.ProvisionSandboxRequest]()
	assert.Nil(t, req.Validate())

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	a := NewActivities(cfg)

	validProvisionOutput := map[string]interface{}{
		clusterIDKey:       "clusterid",
		clusterEndpointKey: "https://k8s.endpoint",
		clusterCAKey:       "b64 encoded ca",
	}

	// Mock activity implementation
	env.OnActivity(a.ApplySandbox, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ApplySandboxRequest) (ApplySandboxResponse, error) {
			assert.Nil(t, pr.validate())

			assert.Equal(t, req.OrgId, pr.OrgID)
			assert.Equal(t, req.AppId, pr.AppID)
			assert.Equal(t, req.InstallId, pr.InstallID)
			assert.Equal(t, cfg.InstallationsBucket, pr.BackendBucketName)
			assert.Equal(t, cfg.InstallationsBucketRegion, pr.BackendBucketRegion)
			assert.Equal(t, cfg.NuonAccessRoleArn, pr.NuonAccessRoleArn)
			assert.Equal(t, cfg.OrgInstanceRoleTemplate, pr.OrgInstanceRoleTemplate)

			return ApplySandboxResponse{Outputs: validProvisionOutput}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionSandbox, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *sandboxv1.ProvisionSandboxResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.Equal(t, validProvisionOutput[clusterIDKey], resp.TerraformOutputs[clusterIDKey])
	require.Equal(t, validProvisionOutput[clusterEndpointKey], resp.TerraformOutputs[clusterEndpointKey])
	require.Equal(t, validProvisionOutput[clusterCAKey], resp.TerraformOutputs[clusterCAKey])
	require.NotNil(t, resp)
}
