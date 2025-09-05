package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type FetchUntornInstallDeploysRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) FetchUntornInstallDeploys(ctx context.Context, req FetchUntornInstallDeploysRequest) ([]*app.InstallDeploy, error) {
	install := app.Install{}

	// can still optimize here with a preload of latest deploy
	res := a.db.WithContext(ctx).
		Preload("InstallComponents").
		First(&install, "id = ?", req.InstallID)

	untornInstallDeploys := make([]*app.InstallDeploy, 0)

	if res.Error == gorm.ErrRecordNotFound {
		return untornInstallDeploys, nil
	}
	if res.Error != nil {
		return untornInstallDeploys, fmt.Errorf("unable to get install: %w", res.Error)
	}

	for _, installCmp := range install.InstallComponents {

		latestDeploy, err := a.getLatestDeploy(ctx, req.InstallID, installCmp.ComponentID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}

		if err != nil {
			return untornInstallDeploys, fmt.Errorf("unable to get latest deploy: %w", err)
		}

		if latestDeploy == nil {
			continue
		} else if !latestDeploy.IsTornDown() {
			untornInstallDeploys = append(untornInstallDeploys, latestDeploy)

		}
	}

	return untornInstallDeploys, nil
}
