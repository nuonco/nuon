package createdeployment

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type activities struct {
	repo repos.DeploymentRepo
	mgr  workflows.DeploymentWorkflowManager
}

func NewActivities(db *gorm.DB, tc tclient.Client) *activities {
	return &activities{
		repo: repos.NewDeploymentRepo(db),
		mgr:  workflows.NewDeploymentWorkflowManager(tc),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerDeploymentJob(ctx context.Context, deploymentID string) (*TriggerJobResponse, error) {
	deploymentUUID, _ := uuid.Parse(deploymentID)

	deployment, err := a.repo.Get(ctx, deploymentUUID)
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
