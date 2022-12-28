package instances

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-generics"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/instances/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/proto"
)

func TestProvisionInstances(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*instancesv1.ProvisionRequest]()
	req.InstallIds = []string{uuid.NewString()}
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}

	// register activities
	act := NewActivities(cfg)
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
			assert.Nil(t, pr.validate())
			assert.Equal(t, req.OrgId, pr.OrgID)
			assert.Equal(t, req.AppId, pr.AppID)
			assert.Equal(t, req.DeploymentId, pr.DeploymentID)
			assert.True(t, proto.Equal(req.Plan, pr.Plan))
			return ProvisionInstanceResponse{WorkflowID: uuid.NewString()}, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ProvisionInstances, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &instancesv1.ProvisionResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
