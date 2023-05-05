package deleteinstall

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type activities struct {
	repo      repos.InstallRepo
	adminRepo repos.AdminRepo
	appRepo   repos.AppRepo
	mgr       workflows.InstallWorkflowManager
}

func NewActivities(db *gorm.DB, tc tclient.Client) *activities {
	return &activities{
		repo:      repos.NewInstallRepo(db),
		adminRepo: repos.NewAdminRepo(db),
		appRepo:   repos.NewAppRepo(db),
		mgr:       workflows.NewInstallWorkflowManager(tc),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerInstallDeprovJob(ctx context.Context, installID string) (*TriggerJobResponse, error) {
	installUUID, _ := uuid.Parse(installID)

	install, err := a.repo.GetDeleted(ctx, installUUID)
	if err != nil {
		return nil, err
	}

	app, err := a.appRepo.Get(ctx, install.AppID)
	if err != nil {
		return nil, err
	}

	sandboxVersion, err := a.adminRepo.GetLatestSandboxVersion(ctx)
	if err != nil {
		return nil, err
	}

	workflow, err := a.mgr.Deprovision(ctx, install, app.OrgID.String(), sandboxVersion)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflow,
	}, nil
}
