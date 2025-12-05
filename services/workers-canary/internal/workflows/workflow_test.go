package workflows

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	sharedactivitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestProvision(t *testing.T) {
	// NOTE(jm): need to reenable later
	return

	v := validator.New()
	cfg := generics.GetFakeObj[workers.Config]()

	wkflow, err := New(v, cfg, nil)
	assert.NoError(t, err)

	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(wkflow.Deprovision)
	act, _ := activities.New(v)
	env.RegisterActivity(act)
	sharedActs, _ := sharedactivities.New(v)
	env.RegisterActivity(sharedActs)

	// Mock activity implementations
	env.OnWorkflow(wkflow.Deprovision, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
			resp := &canaryv1.DeprovisionResponse{}
			assert.Nil(t, r.Validate())
			return resp, nil
		})

	env.OnActivity("SendNotification", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, aReq *sharedactivitiesv1.SendNotificationRequest) (*sharedactivitiesv1.SendNotificationResponse, error) {
			return &sharedactivitiesv1.SendNotificationResponse{}, nil
		})

	env.OnActivity("StartWorkflow", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, aReq *activitiesv1.StartWorkflowRequest) (*activitiesv1.StartWorkflowResponse, error) {
			return &activitiesv1.StartWorkflowResponse{}, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify response
	var resp *canaryv1.ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
