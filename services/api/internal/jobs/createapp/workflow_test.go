package createapp

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
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

	req := &jobsv1.CreateAppRequest{
		AppId: "app-id",
	}

	env.OnActivity(a.TriggerAppJob, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, s string) (*TriggerJobResponse, error) {
			assert.Equal(t, req.AppId, s)
			return &TriggerJobResponse{}, nil
		})

	env.ExecuteWorkflow(wf.CreateApp, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
