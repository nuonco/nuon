package iam

import (
	"context"
	"log"
	"testing"

	"github.com/go-faker/faker/v4"
	iamv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/iam/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/signup/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func getFakeObj[T any]() T {
	var obj T
	err := faker.FakeData(&obj)
	if err != nil {
		log.Fatalf("unable to create fake obj: %s", err)
	}
	return obj
}

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := getFakeObj[workers.Config]()

	wkfl := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(wkfl.Install)

	wf := NewWorkflow(cfg)
	a := NewActivities()

	req := getFakeObj[*iamv1.ProvisionIAMRequest]()

	// Mock activity implementations
	env.OnActivity(a.CreateDeploymentsBucketRole, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r CreateDeploymentsBucketRoleRequest) (CreateDeploymentsBucketRoleResponse, error) {
			resp := CreateDeploymentsBucketRoleResponse{}
			err := r.validate()
			assert.Nil(t, err)
			return resp, nil
		})

	env.ExecuteWorkflow(wf.ProvisionIAM, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *iamv1.ProvisionIAMResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
