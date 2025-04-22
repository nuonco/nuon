package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) EnsureInstallSandbox(ctx context.Context, appID string, installIDs []string) error {
	parentApp := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Installs").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to get app")
	}

	if len(installIDs) < 1 {
		for _, install := range parentApp.Installs {
			installIDs = append(installIDs, install.ID)
		}
	}

	installSandboxes := make([]app.InstallSandbox, 0)
	for _, installID := range installIDs {
		installSandboxes = append(installSandboxes, app.InstallSandbox{
			InstallID: installID,
			Status:    app.InstallSandboxStatusUnknown,
		})
	}

	if len(installSandboxes) < 1 {
		return nil
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&installSandboxes)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to create install sandboxes")
	}

	return nil
}
