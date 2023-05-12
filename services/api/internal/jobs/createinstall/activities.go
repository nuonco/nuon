package createinstall

import (
	"context"

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

	workflow, err := a.mgr.Provision(ctx, install, app.OrgID, sandboxVersion)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflow,
	}, nil
}
