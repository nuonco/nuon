package helpers

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) MarkInstallStateStale(ctx context.Context, installID string) error {
	var is app.InstallState

	err := h.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&is).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // No record to update
		}

		return err
	}

	if res := h.db.WithContext(ctx).Model(&is).Update("stale_at", time.Now()); res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stale_at field")
	}

	return nil
}
