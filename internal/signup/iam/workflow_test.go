package iam

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	iamv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/iam/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/signup/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := generics.GetFakeObj[workers.Config]()

	wkfl := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(wkfl.Install)

	wf := NewWorkflow(cfg)
	a := NewActivities()

	req := generics.GetFakeObj[*iamv1.ProvisionIAMRequest]()

	// Mock activity implementations
	env.OnActivity(a.CreateIAMPolicy, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateIAMPolicyRequest) (CreateIAMPolicyResponse, error) {
			resp := CreateIAMPolicyResponse{
				PolicyArn: "test-policy-arn",
			}
			assert.NoError(t, r.validate())

			return resp, nil
		})

	env.OnActivity(a.CreateIAMRole, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateIAMRoleRequest) (CreateIAMRoleResponse, error) {
			resp := CreateIAMRoleResponse{
				RoleArn: "test-role-arn",
			}
			assert.NoError(t, r.validate())

			return resp, nil
		})

	env.OnActivity(a.CreateIAMRolePolicyAttachment, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateIAMRolePolicyAttachmentRequest) (CreateIAMRolePolicyAttachmentResponse, error) {
			resp := CreateIAMRolePolicyAttachmentResponse{}
			assert.NoError(t, r.validate())
			assert.Equal(t, "test-role-arn", r.RoleArn)
			assert.Equal(t, "test-policy-arn", r.PolicyArn)

			return resp, nil
		})

	env.ExecuteWorkflow(wf.ProvisionIAM, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var resp *iamv1.ProvisionIAMResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.Equal(t, "test-role-arn", resp.DeploymentsRoleArn)
	assert.Equal(t, "test-role-arn", resp.InstallationsRoleArn)
}
