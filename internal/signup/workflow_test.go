package signup

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-common/shortid"
	orgsv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/runner/v1"
	serverv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/server/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/signup/runner"
	"github.com/powertoolsdev/workers-orgs/internal/signup/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := workers.Config{
		WaypointBootstrapTokenNamespace: "default",
		WaypointServerRootDomain:        "testing.nuon.co",
	}
	srv := server.NewWorkflow(cfg)
	run := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(srv.Provision)
	env.RegisterWorkflow(run.Install)

	wf := NewWorkflow(cfg)
	a := NewActivities(nil)

	req := &orgsv1.SignupRequest{OrgId: "00000000-0000-0000-0000-000000000000", Region: "us-west-2"}

	id, err := shortid.ParseString(req.OrgId)
	require.NoError(t, err)

	// Mock activity implementations

	env.OnActivity(a.SendNotification, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, snr SendNotificationRequest) (SendNotificationResponse, error) {
			return SendNotificationResponse{}, nil
		})

	env.OnWorkflow(srv.Provision, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *serverv1.ProvisionRequest) (*serverv1.ProvisionResponse, error) {
			var resp *serverv1.ProvisionResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, id, r.OrgId)
			return resp, nil
		})

	env.OnWorkflow(run.Install, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *runnerv1.InstallRunnerRequest) (*runnerv1.InstallRunnerResponse, error) {
			var resp runnerv1.InstallRunnerResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, id, r.OrgId)
			return &resp, nil
		})

	env.ExecuteWorkflow(wf.Signup, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp orgsv1.SignupResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
}
