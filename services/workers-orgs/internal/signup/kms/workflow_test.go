package kms

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	kmsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/kms/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/roles"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/signup/runner"
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

	req := generics.GetFakeObj[*kmsv1.ProvisionKMSRequest]()

	// Mock activity implementations
	env.OnActivity(a.CreateKMSKey, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateKMSKeyRequest) (CreateKMSKeyResponse, error) {
			resp := CreateKMSKeyResponse{
				KeyArn: "test-policy-arn",
				KeyID:  "abc",
			}
			assert.NoError(t, r.validate())
			assert.Equal(t, cfg.OrgsKMSAccessRoleArn, r.AssumeRoleARN)
			assert.Equal(t, 3, len(r.KeyTags))
			assert.Equal(t, [2]string{"Name", "org/" + req.OrgId}, r.KeyTags[2])

			return resp, nil
		})

	env.OnActivity(a.CreateKMSKeyPolicy, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateKMSKeyPolicyRequest) (CreateKMSKeyPolicyResponse, error) {
			resp := CreateKMSKeyPolicyResponse{}
			assert.NoError(t, r.validate())

			assert.Equal(t, "abc", r.KeyID)
			assert.Equal(t, roles.KeyValuesIAMName(req.OrgId), r.PolicyName)

			expectedPolicy, err := roles.KeyValuesKMSKeyPolicy(req.KeyValuesIamRoleArn)
			assert.NoError(t, err)
			assert.Equal(t, string(expectedPolicy), r.Policy)
			return resp, nil
		})

	env.ExecuteWorkflow(wf.ProvisionKMS, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var resp *kmsv1.ProvisionKMSResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
