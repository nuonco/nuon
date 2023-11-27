package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxRunRequest struct {
	InstallID string             `validate:"required"`
	RunType   app.SandboxRunType `validate:"required"`
}

func (a *Activities) CreateSandboxRun(ctx context.Context, req CreateSandboxRunRequest) (*app.InstallSandboxRun, error) {
	install, err := a.Get(ctx, GetRequest{
		InstallID: req.InstallID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	run := app.InstallSandboxRun{
		OrgID:              install.OrgID,
		RunType:            req.RunType,
		InstallID:          req.InstallID,
		CreatedByID:        install.CreatedByID,
		AppSandboxConfigID: install.AppSandboxConfigID,
	}

	res := a.db.WithContext(ctx).Create(&run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install sandbox run: %w", res.Error)
	}

	return &run, nil
}
