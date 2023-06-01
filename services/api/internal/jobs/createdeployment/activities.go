package createdeployment

import (
	"context"

	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	"gorm.io/gorm"
)

type activities struct {
	repo repos.DeploymentRepo
	mgr  workflows.DeploymentWorkflowManager
}

func NewActivities(db *gorm.DB, workflowsClient workflowsclient.Client) *activities {
	return &activities{
		repo: repos.NewDeploymentRepo(db),
		mgr:  workflows.NewDeploymentWorkflowManager(workflowsClient),
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

	workflowID, err := a.mgr.Start(ctx, deployment)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
