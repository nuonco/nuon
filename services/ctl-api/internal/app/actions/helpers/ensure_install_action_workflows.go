package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) EnsureInstallAction(ctx context.Context, appID string, installIDs []string) error {
	// fetch the parent app's installs and ensure each gets the new component
	parentApp := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Installs").
		Preload("ActionWorkflows").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to get app")
	}

	// if install IDs are not passed in, then update all installs
	if len(installIDs) < 1 {
		for _, install := range parentApp.Installs {
			installIDs = append(installIDs, install.ID)
		}
	}

	// create an install component for all known installs
	installActs := make([]app.InstallActionWorkflow, 0)
	for _, installID := range installIDs {
		for _, actionWorkflow := range parentApp.ActionWorkflows {
			installActs = append(installActs, app.InstallActionWorkflow{
				ActionWorkflowID: actionWorkflow.ID,
				InstallID:        installID,
			})
		}
	}

	if len(installActs) < 1 {
		return nil
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&installActs)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to create install action workflows")
	}

	return nil
}
