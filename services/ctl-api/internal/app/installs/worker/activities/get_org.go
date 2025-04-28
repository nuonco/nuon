package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetOrgRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetOrg(ctx context.Context, req GetOrgRequest) (*app.Org, error) {
	return a.getInstallOrg(ctx, req.InstallID)
}

func (a *Activities) getInstallOrg(ctx context.Context, installID string) (*app.Org, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("App").
		Preload("App.Org").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return install.App.Org, nil
}
