package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	InstallID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Install, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").
		Preload("AppSandboxConfig").
		Preload("AWSAccount").
		First(&install, "id = ?", req.InstallID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
