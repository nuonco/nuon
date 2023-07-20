package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := generics.GetFakeObj[internal.Config]()

	a := NewActivities()

	req := generics.GetFakeObj[*orgsv1.DeprovisionRequest]()
	wkflow := NewWorkflow(cfg)
	iamWkflow := iam.NewWorkflow(cfg)
	env.RegisterWorkflow(iamWkflow.DeprovisionIAM)

	// Mock activity implementation
	env.OnWorkflow(iamWkflow.DeprovisionIAM, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *iamv1.DeprovisionIAMRequest) (*iamv1.DeprovisionIAMResponse, error) {
			assert.NoError(t, r.Validate())
			return &iamv1.DeprovisionIAMResponse{}, nil
		})

	env.OnActivity(a.DestroyNamespace, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, cnr DestroyNamespaceRequest) (DestroyNamespaceResponse, error) {
			require.Equal(t, req.OrgId, cnr.NamespaceName)
			return DestroyNamespaceResponse{}, nil
		})

	env.OnActivity(a.UninstallWaypoint, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, uwr UninstallWaypointRequest) (UninstallWaypointResponse, error) {
			require.Equal(t, fmt.Sprintf("wp-%s", req.OrgId), uwr.ReleaseName)
			return UninstallWaypointResponse{}, nil
		})

	env.ExecuteWorkflow(wkflow.Deprovision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *orgsv1.DeprovisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
