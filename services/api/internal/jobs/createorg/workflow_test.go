package createorg

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

func TestCreateOrgWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	v := validator.New()
	wf := New(v)
	a := NewActivities(nil)

	req := &jobsv1.CreateOrgRequest{
		OrgId: "org-id",
	}

	env.OnActivity(a.TriggerJob, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, s string) (*TriggerJobResponse, error) {
			assert.Equal(t, req.OrgId, s)
			return &TriggerJobResponse{}, nil
		})

	env.ExecuteWorkflow(wf.CreateOrg, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
