package instances

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	provisionv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/instances/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
	tmock "go.temporal.io/sdk/mocks"
)

func TestInstanceProvisioner_startWorkflow(t *testing.T) {
	tc := &tmock.Client{}
	ip := instanceProvisioner{}
	req := generics.GetFakeObj[*provisionv1.ProvisionRequest]()

	expectedOpts := tclient.StartWorkflowOptions{TaskQueue: "instance",
		Memo: map[string]interface{}{
			"org-id":        req.OrgId,
			"app-id":        req.AppId,
			"deployment-id": req.DeploymentId,
			"install-id":    req.InstallId,
		},
	}
	workflowRun := &tmock.WorkflowRun{}

	workflowRun.On("Get", mock.Anything, mock.Anything).Return(nil)
	tc.On("ExecuteWorkflow", mock.Anything, expectedOpts, "Provision", req).Return(workflowRun, nil)

	err := ip.startWorkflow(context.Background(), tc, req)
	assert.Nil(t, err)
}
