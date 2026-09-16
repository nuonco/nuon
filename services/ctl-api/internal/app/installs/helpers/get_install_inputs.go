package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

// getInstallInputs gets the inputs and their current values for an install from the DB.
func (h *Helpers) getInstallInputs(ctx context.Context, installID string) ([]app.InstallInputs, error) {
	var install app.Install
	res := h.db.WithContext(ctx).
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC"))
		}).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install inputs: %w", res.Error)
	}

	return install.InstallInputs, nil
}
