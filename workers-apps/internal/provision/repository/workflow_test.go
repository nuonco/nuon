package repository

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	repov1 "github.com/powertoolsdev/protos/workflows/generated/types/apps/v1/repository/v1"
	"github.com/powertoolsdev/workers-apps/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestRunner(t *testing.T) {
	cfg := generics.GetFakeObj[internal.Config]()
	req := generics.GetFakeObj[*repov1.ProvisionRepositoryRequest]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities()

	env.OnActivity(a.CreateRepository, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crReq CreateRepositoryRequest) (CreateRepositoryResponse, error) {
			assert.Nil(t, crReq.validate())
			return CreateRepositoryResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionRepository, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	resp := &repov1.ProvisionRepositoryResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
