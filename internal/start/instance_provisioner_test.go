package start

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-waypoint"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func TestInstanceProvisioner_startWorkflow(t *testing.T) {
	tc := &tmock.Client{}
	ip := instanceProvisioner{}
	req := getFakeObj[*deploymentsv1.StartRequest]()
	pReq := ProvisionInstanceRequest{
		AppID:        req.AppId,
		OrgID:        req.OrgId,
		InstallID:    uuid.NewString(),
		DeploymentID: req.DeploymentId,
		Component:    waypoint.Component{},
	}

	expectedOpts := tclient.StartWorkflowOptions{TaskQueue: "instance"}
	workflowRun := &tmock.WorkflowRun{}
	workflowRun.On("GetID").Return("abc")
	tc.On("ExecuteWorkflow", mock.Anything, expectedOpts, "Provision", pReq).Return(workflowRun, nil)

	workflowID, err := ip.startWorkflow(context.Background(), tc, pReq)
	assert.Nil(t, err)
	assert.Equal(t, workflowRun.GetID(), workflowID)
}
