package workflows

import (
	"context"
	"fmt"
	"time"

	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultDeployTimeout time.Duration = time.Second * 10
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_deployments.go -source=deployments.go -package=workflows
func NewDeploymentWorkflowManager(workflowsClient workflowsclient.Client) *deploymentWorkflowManager {
	return &deploymentWorkflowManager{workflowsClient}
}

type deploymentWorkflowManager struct {
	workflowsClient workflowsclient.Client
}

type DeploymentWorkflowManager interface {
	Start(context.Context, *models.Deployment) (string, error)
}

var _ DeploymentWorkflowManager = (*deploymentWorkflowManager)(nil)

func (d *deploymentWorkflowManager) Start(ctx context.Context, deployment *models.Deployment) (string, error) {
	app := deployment.Component.App
	orgID, appID, componentID, deploymentID := app.OrgID, app.ID, deployment.Component.ID, deployment.ID
	var compConf componentv1.Component
	if deployment.Component.Config != nil {
		if err := protojson.Unmarshal([]byte(deployment.Component.Config.String()), &compConf); err != nil {
			return "", fmt.Errorf("failed to unmarshal DB JSON: %w", err)
		}
	}
	compConf.Id = componentID
	compConf.DeployCfg.Timeout = durationpb.New(defaultDeployTimeout)

	req := &deploymentsv1.StartRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
		InstallIds:   make([]string, len(app.Installs)),
		Component:    &compConf,
	}
	for idx, install := range app.Installs {
		req.InstallIds[idx] = install.ID
	}

	workflowID, err := d.workflowsClient.TriggerDeploymentStart(ctx, req)
	if err != nil {
		return "", fmt.Errorf("unable to start deployment: %w", err)
	}

	return workflowID, nil
}
