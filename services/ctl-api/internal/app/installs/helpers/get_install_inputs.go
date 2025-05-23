package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"gorm.io/gorm"
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
