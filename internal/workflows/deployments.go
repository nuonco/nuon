package workflows

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/go-common/shortid"
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
	app := deployment.Component.App
	shortIDs, err := shortid.ParseStrings(app.OrgID.String(), app.ID.String(), deployment.Component.ID.String(), deployment.ID.String())
	if err != nil {
		return fmt.Errorf("unable to parse ids to shortids: %w", err)
	}
	orgID, appID, componentID, deploymentID := shortIDs[0], shortIDs[1], shortIDs[2], shortIDs[3]

	opts := tclient.StartWorkflowOptions{
		TaskQueue: "deployment",
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":        orgID,
			"app-id":        appID,
			"deployment-id": deploymentID,
			"component-id":  componentID,
			"started-by":    "api",
		},
	}
	req := &deploymentsv1.StartRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
		InstallIds:   make([]string, len(app.Installs)),

		// TODO(jm): make this an actual component, instead of a hard coded component type
		Component: &componentv1.Component{
			Id:   componentID,
			Name: deployment.Component.Name,
		},
	}
	for idx, install := range app.Installs {
		installID := shortid.ParseUUID(install.ID)
		req.InstallIds[idx] = installID
	}

	_, err = d.tc.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}
