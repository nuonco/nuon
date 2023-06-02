package createinstall

import (
	"context"

	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/repos"

	"gorm.io/gorm"
)

type activities struct {
	repo      repos.InstallRepo
	adminRepo repos.AdminRepo
	appRepo   repos.AppRepo
	wfc       wfc.Client
}

func NewActivities(db *gorm.DB, workflowsClient wfc.Client) *activities {
	return &activities{
		repo:      repos.NewInstallRepo(db),
		adminRepo: repos.NewAdminRepo(db),
		appRepo:   repos.NewAppRepo(db),
		wfc:       workflowsClient,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerInstallJob(ctx context.Context, installID string) (*TriggerJobResponse, error) {
	install, err := a.repo.Get(ctx, installID)
	if err != nil {
		return nil, err
	}

	// NOTE(jm): this is a hack, we should figure out how to grab the app id without having to pass the appRepo in
	// here etc. Maybe a method on the installRepo, called GetAppOrgID?
	app, err := a.appRepo.Get(ctx, install.AppID)
	if err != nil {
		return nil, err
	}

	sandboxVersion, err := a.adminRepo.GetLatestSandboxVersion(ctx)
	if err != nil {
		return nil, err
	}

	req := install.ToProvisionRequest(app.OrgID, sandboxVersion)

	workflow, err := a.wfc.TriggerInstallProvision(ctx, req)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflow,
	}, nil
}
