package migrations

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (m *Migrations) Migration137InstallAppBranchConnections(ctx context.Context, db *gorm.DB) error {
	if !db.Migrator().HasColumn("installs", "app_branch_id") {
		return nil
	}

	var installs []struct {
		ID          string
		OrgID       string
		CreatedByID string
		AppBranchID string
	}
	if err := db.WithContext(ctx).
		Table("installs").
		Select("id, org_id, created_by_id, app_branch_id").
		Where("app_branch_id IS NOT NULL AND app_branch_id != ''").
		Where(`NOT EXISTS (
			SELECT 1
			FROM install_app_branch_connections
			WHERE install_app_branch_connections.install_id = installs.id
			  AND install_app_branch_connections.active = TRUE
			  AND install_app_branch_connections.deleted_at = 0
		)`).
		Scan(&installs).Error; err != nil {
		return fmt.Errorf("unable to load installs without app branch connections: %w", err)
	}

	now := time.Now()
	for _, install := range installs {
		connection := map[string]any{
			"id":            domains.NewInstallAppBranchConnectionID(),
			"created_by_id": install.CreatedByID,
			"created_at":    now,
			"updated_at":    now,
			"deleted_at":    0,
			"org_id":        install.OrgID,
			"install_id":    install.ID,
			"app_branch_id": install.AppBranchID,
			"active":        true,
			"activated_at":  now,
		}
		if err := db.WithContext(ctx).
			Model(&app.InstallAppBranchConnection{}).
			Create(connection).Error; err != nil {
			return fmt.Errorf("unable to backfill install %s app branch connection: %w", install.ID, err)
		}
	}

	return nil
}
