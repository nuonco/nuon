package sandbox

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	workers "github.com/powertoolsdev/workers-installs/internal"
)

func newFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func getFakeProvisionRequest() ProvisionRequest {
	fkr := faker.New()
	var req ProvisionRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestProvisionSandbox(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	assert.Nil(t, cfg.Validate())

	a := NewActivities(cfg)

	req := getFakeProvisionRequest()
	validProvisionOutput := map[string]string{
		clusterIDKey:       "clusterid",
		clusterEndpointKey: "https://k8s.endpoint",
		clusterCAKey:       "b64 encoded ca",
	}

	// Mock activity implementation
	env.OnActivity(a.ApplySandbox, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ApplySandboxRequest) (ApplySandboxResponse, error) {
			assert.Nil(t, pr.validate())

			assert.Equal(t, req.OrgID, pr.OrgID)
			assert.Equal(t, req.AppID, pr.AppID)
			assert.Equal(t, req.InstallID, pr.InstallID)
			assert.Equal(t, req.AccountSettings, pr.AccountSettings)
			assert.Equal(t, req.SandboxSettings, pr.SandboxSettings)
			return ApplySandboxResponse{Outputs: validProvisionOutput}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionSandbox, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.Equal(t, validProvisionOutput, resp.TerraformOutputs)
	require.NotNil(t, resp)
}
