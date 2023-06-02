package createapp

import (
	"context"

	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"gorm.io/gorm"
)

type activities struct {
	repo            repos.AppRepo
	workflowsClient workflowsclient.Client
}

func NewActivities(db *gorm.DB, workflowsClient workflowsclient.Client) *activities {
	return &activities{
		repo:            repos.NewAppRepo(db),
		workflowsClient: workflowsClient,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerAppJob(ctx context.Context, appID string) (*TriggerJobResponse, error) {
	app, err := a.repo.Get(ctx, appID)
	if err != nil {
		return nil, err
	}

	workflowID, err := a.workflowsClient.TriggerAppProvision(ctx, app.ToProvisionRequest())
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{WorkflowID: workflowID}, nil
}
