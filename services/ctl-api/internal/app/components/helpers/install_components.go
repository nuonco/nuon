package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

func (h *Helpers) EnsureInstallComponents(ctx context.Context, appID string, installIDs []string) error {
	// fetch the parent app's installs and ensure each gets the new component
	parentApp := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Installs").
		Preload("Components").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return fmt.Errorf("unable to get app: %w", res.Error)
	}

	// if install IDs are not passed in, then update all installs
	if len(installIDs) < 1 {
		for _, install := range parentApp.Installs {
			installIDs = append(installIDs, install.ID)
		}
	}

	// create an install component for all known installs
	var installCmps = make([]app.InstallComponent, 0)
	for _, installID := range installIDs {
		for _, component := range parentApp.Components {
			installCmps = append(installCmps, app.InstallComponent{
				ComponentID: component.ID,
				InstallID:   installID,
			})
		}
	}

	if len(installCmps) < 1 {
		return nil
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&installCmps)
	if res.Error != nil {
		return fmt.Errorf("unable to create install components: %w", res.Error)
	}

	return nil
}
