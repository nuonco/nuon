package repository

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

func getFakeProvisionRepositoryRequest() ProvisionRepositoryRequest {
	fkr := faker.New()
	var req ProvisionRepositoryRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestRunner(t *testing.T) {
	cfg := getFakeConfig()
	req := getFakeProvisionRepositoryRequest()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities()

	env.OnActivity(a.CreateRepository, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crReq CreateRepositoryRequest) (CreateRepositoryResponse, error) {
			var resp CreateRepositoryResponse
			assert.Nil(t, crReq.validate())

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionRepository, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionRepositoryResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
