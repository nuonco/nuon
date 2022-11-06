package teardown

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	a := NewActivities()

	req := TeardownRequest{OrgID: "00000000-0000-0000-0000-000000000000", Region: "us-east-2"}

	id, err := shortid.ParseString(req.OrgID)
	require.NoError(t, err)

	// Mock activity implementation
	env.OnActivity(a.DestroyNamespace, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, cnr DestroyNamespaceRequest) (DestroyNamespaceResponse, error) {
			require.Equal(t, id, cnr.NamespaceName)
			return DestroyNamespaceResponse{}, nil
		})

	env.OnActivity(a.UninstallWaypoint, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, uwr UninstallWaypointRequest) (UninstallWaypointResponse, error) {
			require.Equal(t, fmt.Sprintf("wp-%s", id), uwr.ReleaseName)
			return UninstallWaypointResponse{}, nil
		})

	env.ExecuteWorkflow(Teardown, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp TeardownResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
