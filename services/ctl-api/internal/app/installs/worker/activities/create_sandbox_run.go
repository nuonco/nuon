package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateSandboxRunRequest struct {
	InstallID string             `validate:"required"`
	RunType   app.SandboxRunType `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateSandboxRun(ctx context.Context, req CreateSandboxRunRequest) (*app.InstallSandboxRun, error) {
	install, err := a.Get(ctx, GetRequest{
		InstallID: req.InstallID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var status app.SandboxRunStatus
	switch req.RunType {
	case app.SandboxRunTypeProvision:
		status = app.SandboxRunStatusProvisioning
	case app.SandboxRunTypeReprovision:
		status = app.SandboxRunStatusReprovisioning
	case app.SandboxRunTypeDeprovision:
		status = app.SandboxRunStatusDeprovisioning
	default:
		return nil, fmt.Errorf("invalid run type: %s", req.RunType)
	}

	run := app.InstallSandboxRun{
		OrgID:              install.OrgID,
		RunType:            req.RunType,
		InstallID:          req.InstallID,
		CreatedByID:        install.CreatedByID,
		AppSandboxConfigID: install.AppSandboxConfigID,
		Status:             status,
	}

	res := a.db.WithContext(ctx).Create(&run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install sandbox run: %w", res.Error)
	}

	// create terraform workspace
	workspace := app.TerraformWorkspace{
		OrgID:     install.OrgID,
		OwnerID:   run.ID,
		OwnerType: app.TerraformWorkspaceOwnerInstallSandboxRun,
	}

	res = a.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&workspace)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create terraform workspace: %w", res.Error)
	}

	return &run, nil
}
