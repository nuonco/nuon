package deleteinstall

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateAppWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	v := validator.New()
	wf := New(v)
	a := NewActivities(nil, nil)

	req := &jobsv1.DeleteInstallRequest{
		InstallId: "install-id",
	}

	sharedActs, _ := sharedactivities.New(v)
	env.RegisterActivity(sharedActs)

	env.OnActivity(a.TriggerInstallDeprovision, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, s string) (*TriggerJobResponse, error) {
			assert.Equal(t, req.InstallId, s)
			return &TriggerJobResponse{}, nil
		})

	env.OnActivity(sharedActs.PollWorkflow, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *activitiesv1.PollWorkflowRequest) (*activitiesv1.PollWorkflowResponse, error) {
			return &activitiesv1.PollWorkflowResponse{}, nil
		})

	env.ExecuteWorkflow(wf.DeleteInstall, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
