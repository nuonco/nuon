package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
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
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get install for runnner")
	}

	return &install, nil
}
