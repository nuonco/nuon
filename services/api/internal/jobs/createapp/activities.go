package createapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type activities struct {
	repo repos.AppRepo
	mgr  workflows.AppWorkflowManager
}

func NewActivities(db *gorm.DB, tc tclient.Client) *activities {
	return &activities{
		repo: repos.NewAppRepo(db),
		mgr:  workflows.NewAppWorkflowManager(tc),
	}
}

type TriggerJobResponse struct{}

func (a *activities) TriggerAppJob(ctx context.Context, appID string) (*TriggerJobResponse, error) {
	appUUID, _ := uuid.Parse(appID)

	finalApp, err := a.repo.Get(ctx, appUUID)
	if err != nil {
		return nil, err
	}

	err = a.mgr.Provision(ctx, finalApp)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{}, nil
}
