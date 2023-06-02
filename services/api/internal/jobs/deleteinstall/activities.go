package deleteinstall

import (
	"context"

	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"gorm.io/gorm"
)

type activities struct {
	repo      repos.InstallRepo
	adminRepo repos.AdminRepo
	appRepo   repos.AppRepo
	wfc       workflowsclient.Client
}

func NewActivities(db *gorm.DB, wfc workflowsclient.Client) *activities {
	return &activities{
		repo:      repos.NewInstallRepo(db),
		adminRepo: repos.NewAdminRepo(db),
		appRepo:   repos.NewAppRepo(db),
		wfc:       wfc,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerInstallDeprovision(ctx context.Context, installID string) (*TriggerJobResponse, error) {
	install, err := a.repo.GetDeleted(ctx, installID)
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

	req := install.ToDeprovisionRequest(app.OrgID, sandboxVersion)
	workflowID, err := a.wfc.TriggerInstallDeprovision(ctx, req)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
