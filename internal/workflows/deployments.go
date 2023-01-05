package workflows

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_deployments.go -source=deployments.go -package=workflows
func NewDeploymentWorkflowManager(tclient temporalClient) *deploymentWorkflowManager {
	return &deploymentWorkflowManager{
		tc: tclient,
	}
}

type deploymentWorkflowManager struct {
	tc temporalClient
}

type DeploymentWorkflowManager interface {
	Start(context.Context, *models.Deployment) error
}

var _ DeploymentWorkflowManager = (*deploymentWorkflowManager)(nil)

func (d *deploymentWorkflowManager) Start(ctx context.Context, deployment *models.Deployment) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "deployment",
	}

	app := deployment.Component.App
	req := &deploymentsv1.StartRequest{
		OrgId:        app.OrgID.String(),
		AppId:        app.ID.String(),
		DeploymentId: deployment.ID.String(),
		InstallIds:   make([]string, len(app.Installs)),

		// TODO(jm): make this an actual component, instead of a hard coded component type
		Component: &componentv1.Component{
			Id:   uuid.NewString(),
			Name: deployment.Component.Name,
		},
	}
	for idx, install := range app.Installs {
		req.InstallIds[idx] = install.ID.String()
	}

	_, err := d.tc.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}
