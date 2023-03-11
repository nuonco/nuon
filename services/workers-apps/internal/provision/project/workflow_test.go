package project

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	projectv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/apps/v1/project/v1"
	"github.com/powertoolsdev/mono/services/workers-apps/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestWorkflow(t *testing.T) {
	cfg := generics.GetFakeObj[internal.Config]()
	req := generics.GetFakeObj[*projectv1.ProvisionProjectRequest]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities()

	env.OnActivity(act.PingWaypointServer, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwsReq PingWaypointServerRequest) (PingWaypointServerResponse, error) {
			var resp PingWaypointServerResponse
			assert.Nil(t, pwsReq.validate())

			return resp, nil
		})

	env.OnActivity(act.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crReq CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			var resp CreateWaypointProjectResponse
			assert.Nil(t, crReq.validate())

			return resp, nil
		})

	env.OnActivity(act.UpsertWaypointWorkspace, mock.Anything, mock.Anything).
		Return(func(_ context.Context, uwwReq UpsertWaypointWorkspaceRequest) (UpsertWaypointWorkspaceResponse, error) {
			var resp UpsertWaypointWorkspaceResponse
			assert.Nil(t, uwwReq.validate())

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionProject, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *projectv1.ProvisionProjectResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
