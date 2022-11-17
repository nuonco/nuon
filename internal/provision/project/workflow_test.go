package project

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func getFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func getFakeProvisionProjectRequest() ProvisionProjectRequest {
	fkr := faker.New()
	var req ProvisionProjectRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestWorkflow(t *testing.T) {
	cfg := getFakeConfig()
	req := getFakeProvisionProjectRequest()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities()

	env.OnActivity(act.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crReq CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			var resp CreateWaypointProjectResponse
			assert.Nil(t, crReq.validate())

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionProject, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionProjectResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
