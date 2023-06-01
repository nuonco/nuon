package deleteinstall

import (
	"context"

	pkgWorkflows "github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	"gorm.io/gorm"
)

type activities struct {
	repo      repos.InstallRepo
	adminRepo repos.AdminRepo
	appRepo   repos.AppRepo
	mgr       workflows.InstallWorkflowManager
}

func NewActivities(db *gorm.DB, workflowsClient pkgWorkflows.Client) *activities {
	return &activities{
		repo:      repos.NewInstallRepo(db),
		adminRepo: repos.NewAdminRepo(db),
		appRepo:   repos.NewAppRepo(db),
		mgr:       workflows.NewInstallWorkflowManager(workflowsClient),
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

	workflow, err := a.mgr.Deprovision(ctx, install, app.OrgID, sandboxVersion)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflow,
	}, nil
}
