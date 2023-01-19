package instances

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	provisionv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func TestInstanceProvisioner_startWorkflow(t *testing.T) {
	tc := &tmock.Client{}
	ip := instanceProvisioner{}
	req := generics.GetFakeObj[*provisionv1.ProvisionRequest]()

	expectedOpts := tclient.StartWorkflowOptions{TaskQueue: "instance"}
	workflowRun := &tmock.WorkflowRun{}
	workflowRun.On("GetID").Return("abc")
	tc.On("ExecuteWorkflow", mock.Anything, expectedOpts, "Provision", req).Return(workflowRun, nil)

	workflowID, err := ip.startWorkflow(context.Background(), tc, req)
	assert.Nil(t, err)
	assert.Equal(t, workflowRun.GetID(), workflowID)
}
