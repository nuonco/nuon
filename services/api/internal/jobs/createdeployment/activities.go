package createdeployment

import (
	"context"
	"fmt"

	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"gorm.io/gorm"
)

type activities struct {
	repo repos.DeploymentRepo
	wfc  workflowsclient.Client
}

func NewActivities(db *gorm.DB, workflowsClient workflowsclient.Client) *activities {
	return &activities{
		repo: repos.NewDeploymentRepo(db),
		wfc:  workflowsClient,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerDeploymentJob(ctx context.Context, deploymentID string) (*TriggerJobResponse, error) {
	deployment, err := a.repo.Get(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	startReq, err := deployment.ToStartRequest()
	if err != nil {
		return nil, fmt.Errorf("error triggering deployment job: %w", err)
	}

	workflowID, err := a.wfc.TriggerDeploymentStart(ctx, startReq)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
