package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxRunRequest struct {
	InstallID  string             `validate:"required"`
	RunType    app.SandboxRunType `validate:"required"`
	WorkflowID string             `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateSandboxRun(ctx context.Context, req CreateSandboxRunRequest) (*app.InstallSandboxRun, error) {
	install, err := a.Get(ctx, GetRequest{
		InstallID: req.InstallID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	appCfg, err := a.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get full app config")
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
		AppSandboxConfigID: appCfg.SandboxConfig.ID,
		Status:             status,
		InstallWorkflowID:  generics.ToPtr(req.WorkflowID),
	}

	resCreateRun := a.db.WithContext(ctx).Create(&run)
	if resCreateRun.Error != nil {
		return nil, fmt.Errorf("unable to create install sandbox run: %w", resCreateRun.Error)
	}

	return &run, nil
}
