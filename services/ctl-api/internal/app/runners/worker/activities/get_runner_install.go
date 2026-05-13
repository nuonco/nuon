package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunnerInstallRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetRunnerInstall(ctx context.Context, req GetRunnerInstallRequest) (*app.Install, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install for runner: %w", res.Error)
	}

	return &install, nil
}
