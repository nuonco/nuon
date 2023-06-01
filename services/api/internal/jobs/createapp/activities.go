package createapp

import (
	"context"

	pkgWorkflows "github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	"gorm.io/gorm"
)

type activities struct {
	repo repos.AppRepo
	mgr  workflows.AppWorkflowManager
}

func NewActivities(db *gorm.DB, workflowsClient pkgWorkflows.Client) *activities {
	return &activities{
		repo: repos.NewAppRepo(db),
		mgr:  workflows.NewAppWorkflowManager(workflowsClient),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerAppJob(ctx context.Context, appID string) (*TriggerJobResponse, error) {
	finalApp, err := a.repo.Get(ctx, appID)
	if err != nil {
		return nil, err
	}

	wrkflwID, err := a.mgr.Provision(ctx, finalApp)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{WorkflowID: wrkflwID}, nil
}
