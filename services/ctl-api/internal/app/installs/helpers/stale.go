package helpers

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (h *Helpers) MarkInstallStateStale(ctx context.Context, installID string) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := LockInstallInputs(ctx, tx, installID); err != nil {
			return err
		}
		var is app.InstallState
		err := tx.Where(app.InstallState{InstallID: installID}).Order("created_at DESC").First(&is).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Model(&is).Update("stale_at", time.Now()).Error; err != nil {
			return errors.Wrap(err, "unable to update stale_at field")
		}
		return nil
	})
}

// MarkInstallStatePartialsStale marks the latest install state stale and appends the given partials
// to its stale_partials list (unioned + deduped).
func (h *Helpers) MarkInstallStatePartialsStale(ctx context.Context, db *gorm.DB, installID string, partials ...pkgstate.PartialName) error {
	if len(partials) == 0 {
		return nil
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := LockInstallInputs(ctx, tx, installID); err != nil {
			return err
		}
		var is app.InstallState
		err := tx.Where(app.InstallState{InstallID: installID}).Order("created_at DESC").First(&is).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		set := make(map[pkgstate.PartialName]struct{}, len(is.StalePartials)+len(partials))
		for _, p := range is.StalePartials {
			set[p] = struct{}{}
		}
		for _, p := range partials {
			set[p] = struct{}{}
		}

		merged := make([]pkgstate.PartialName, 0, len(set))
		for _, p := range pkgstate.AllPartials {
			if _, ok := set[p]; ok {
				merged = append(merged, p)
			}
		}

		if err := tx.Model(&is).
			Select("stale_at", "stale_partials").
			Updates(app.InstallState{
				StaleAt:       generics.NewNullTime(time.Now()),
				StalePartials: merged,
			}).Error; err != nil {
			return errors.Wrap(err, "unable to update stale_partials field")
		}
		return nil
	})
}
