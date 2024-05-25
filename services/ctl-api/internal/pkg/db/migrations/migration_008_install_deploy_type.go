package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration008InstallDeployType(ctx context.Context) error {
	var installDeploys []*app.InstallDeploy
	res := a.db.WithContext(ctx).
		Find(&installDeploys)

	if res.Error != nil {
		return res.Error
	}

	for _, installDeploy := range installDeploys {
		if installDeploy.Type != app.InstallDeployType("") {
			continue
		}

		if installDeploy.ComponentReleaseStepID != nil {
			installDeploy.Type = app.InstallDeployTypeRelease
		} else {
			installDeploy.Type = app.InstallDeployTypeInstall
		}

		a.l.Info("example migration - gorm update")
	}

	return nil
}
