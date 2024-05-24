package provision

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	kmsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/kms/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/runner/v1"
	serverv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/server/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/iam"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/kms"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/runner"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := workers.Config{
		WaypointBootstrapTokenNamespace: "default",
		WaypointServerRootDomain:        "testing.nuon.co",
	}
	srv := server.NewWorkflow(cfg)
	iamer := iam.NewWorkflow(cfg)
	kmser := kms.NewWorkflow(cfg)
	run := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(srv.ProvisionServer)
	env.RegisterWorkflow(iamer.ProvisionIAM)
	env.RegisterWorkflow(kmser.ProvisionKMS)
	env.RegisterWorkflow(run.ProvisionRunner)

	wf := NewWorkflow(cfg)
	a := NewActivities()

	req := generics.GetFakeObj[*orgsv1.ProvisionRequest]()
	iamResp := generics.GetFakeObj[*iamv1.ProvisionIAMResponse]()
	kmsResp := generics.GetFakeObj[*kmsv1.ProvisionKMSResponse]()
	serverResp := generics.GetFakeObj[*serverv1.ProvisionServerResponse]()

	// Mock activity implementations
	env.OnActivity(a.StartSignupRequest, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			return &sharedv1.StartActivityResponse{}, nil
		})

	env.OnActivity(a.FinishSignupRequest, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			return &sharedv1.FinishActivityResponse{}, nil
		})

	env.OnWorkflow(kmser.ProvisionKMS, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *kmsv1.ProvisionKMSRequest) (*kmsv1.ProvisionKMSResponse, error) {
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, iamResp.SecretsRoleArn, r.SecretsIamRoleArn)
			return kmsResp, nil
		})

	env.OnWorkflow(iamer.ProvisionIAM, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *iamv1.ProvisionIAMRequest) (*iamv1.ProvisionIAMResponse, error) {
			assert.Nil(t, r.Validate())
			return iamResp, nil
		})

	env.OnWorkflow(srv.ProvisionServer, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *serverv1.ProvisionServerRequest) (*serverv1.ProvisionServerResponse, error) {
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			return serverResp, nil
		})

	env.OnWorkflow(run.ProvisionRunner, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
			var resp runnerv1.ProvisionRunnerResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			return &resp, nil
		})

	env.ExecuteWorkflow(wf.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// assert response
	var resp orgsv1.ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	assert.True(t, proto.Equal(resp.IamRoles, iamResp))
	assert.True(t, proto.Equal(resp.Server, serverResp))
}
