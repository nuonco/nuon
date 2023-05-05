package workflows

import (
	"context"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	tclient "go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultDeployTimeout time.Duration = time.Second * 10
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
	Start(context.Context, *models.Deployment) (string, error)
}

var _ DeploymentWorkflowManager = (*deploymentWorkflowManager)(nil)

func (d *deploymentWorkflowManager) Start(ctx context.Context, deployment *models.Deployment) (string, error) {
	app := deployment.Component.App
	shortIDs, err := shortid.ParseStrings(app.OrgID.String(), app.ID.String(), deployment.Component.ID.String(), deployment.ID.String())
	if err != nil {
		return "", fmt.Errorf("unable to parse ids to shortids: %w", err)
	}
	orgID, appID, componentID, deploymentID := shortIDs[0], shortIDs[1], shortIDs[2], shortIDs[3]

	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":        orgID,
			"app-id":        appID,
			"deployment-id": deploymentID,
			"component-id":  componentID,
			"started-by":    "api",
		},
	}

	var compConf componentv1.Component
	if deployment.Component.Config != nil {
		if err = protojson.Unmarshal([]byte(deployment.Component.Config.String()), &compConf); err != nil {
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
		installID := shortid.ParseUUID(install.ID)
		req.InstallIds[idx] = installID
	}

	workflow, err := d.tc.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return "", fmt.Errorf("unable to start deployment: %w", err)
	}

	return workflow.GetID(), nil
}
