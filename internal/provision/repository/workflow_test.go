package repository

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/workers-apps/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestRunner(t *testing.T) {
	cfg := generics.GetFakeObj[internal.Config]()
	req := generics.GetFakeObj[ProvisionRepositoryRequest]()

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
