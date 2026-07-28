package migrations

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func (m *Migrations) Migration116InstallAppConfigVersionMetadataBackfill(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS install_config_updates`).Error; err != nil {
		return fmt.Errorf("unable to drop old install_config_updates table: %w", err)
	}

	if err := db.WithContext(ctx).AutoMigrate(&app.InstallAppConfigVersion{}); err != nil {
		return fmt.Errorf("unable to create install_app_config_versions table: %w", err)
	}

	// installs whose creator is gone would fail the accounts FK, so fall back instead of
	// skipping them
	systemCtx, err := m.systemCtx(ctx)
	if err != nil {
		return err
	}
	fallbackCreatedByID := keys.CreatedByIDFromContext(systemCtx)

	var installs []struct {
		ID          string
		OrgID       string
		AppConfigID string
		CreatedByID string
	}
	if err := db.WithContext(ctx).Raw(`
		SELECT i.id, i.org_id, i.app_config_id,
		       CASE WHEN EXISTS (SELECT 1 FROM accounts a WHERE a.id = i.created_by_id)
		            THEN i.created_by_id ELSE '' END AS created_by_id
		FROM installs i
		WHERE i.app_config_id != ''
		  AND i.deleted_at = 0
		  AND NOT EXISTS (
			SELECT 1 FROM install_app_config_versions icu
			WHERE icu.install_id = i.id AND icu.deleted_at = 0
		  )
	`).Scan(&installs).Error; err != nil {
		return fmt.Errorf("unable to query installs for backfill: %w", err)
	}

	now := time.Now()
	for _, inst := range installs {
		createdByID := inst.CreatedByID
		if createdByID == "" {
			createdByID = fallbackCreatedByID
		}

		version := app.InstallAppConfigVersion{
			ID:             domains.NewInstallAppConfigVersionID(),
			OrgID:          inst.OrgID,
			InstallID:      inst.ID,
			NewAppConfigID: inst.AppConfigID,
			CreatedByID:    createdByID,
			CreatedAt:      now,
			UpdatedAt:      now,
			Status:         app.CompositeStatus{Status: app.StatusSuccess},
			Metadata:       map[string]string{"source": "backfill"},
		}
		// backfilled rows have no previous config, and '' fails the app_configs FK
		if err := db.WithContext(ctx).Omit("OldAppConfigID").Create(&version).Error; err != nil {
			return fmt.Errorf("unable to backfill install %s: %w", inst.ID, err)
		}
	}

	return nil
}
