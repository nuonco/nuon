package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) EnsureInstallRunbooks(ctx context.Context, appID string, installIDs []string) error {
	parentApp := app.App{}
	res := h.db.WithContext(ctx).
		Preload("Installs").
		Preload("Runbooks").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return fmt.Errorf("unable to get app: %w", res.Error)
	}

	if len(installIDs) < 1 {
		for _, install := range parentApp.Installs {
			installIDs = append(installIDs, install.ID)
		}
	}

	installRunbooks := make([]app.InstallRunbook, 0)
	for _, installID := range installIDs {
		for _, runbook := range parentApp.Runbooks {
			installRunbooks = append(installRunbooks, app.InstallRunbook{
				RunbookID: runbook.ID,
				InstallID: installID,
			})
		}
	}

	if len(installRunbooks) < 1 {
		return nil
	}

	res = h.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&installRunbooks)
	if res.Error != nil {
		return fmt.Errorf("unable to create install runbooks: %w", res.Error)
	}

	return nil
}
