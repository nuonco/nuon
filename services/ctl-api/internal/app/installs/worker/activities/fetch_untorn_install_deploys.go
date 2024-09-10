package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type FetchUntornInstallDeploysRequest struct {
	InstallID string `json:"install_id"`
}

// @await-gen
func (a *Activities) FetchUntornInstallDeploys(ctx context.Context, req FetchUntornInstallDeploysRequest) ([]string, error) {
	install := app.Install{}

	res := a.db.WithContext(ctx).
		Preload("InstallComponents").
		// can still optimize here with a preload of latest deploy
		First(&install, "id = ?", req.InstallID)

	untornCmpIDs := make([]string, 0)

	if res.Error == gorm.ErrRecordNotFound {
		return untornCmpIDs, nil
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	for _, installCmp := range install.InstallComponents {

		latestDeploy, err := a.getLatestDeploy(ctx, req.InstallID, installCmp.ComponentID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
	
		if err != nil {
			return nil, fmt.Errorf("unable to get latest deploy: %w", err)
		}

		if latestDeploy == nil {
			continue
		}

		if latestDeploy != nil {
			deployTornDown := latestDeploy.Status == app.InstallDeployStatusOK && latestDeploy.Type == app.InstallDeployTypeTeardown
			if !deployTornDown {
				untornCmpIDs = append(untornCmpIDs, installCmp.ComponentID)
			}
		}
	}

	return untornCmpIDs, nil
}
