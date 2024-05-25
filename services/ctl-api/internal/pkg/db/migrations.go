package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (a *AutoMigrate) isMigrationApplied(ctx context.Context, name string) (bool, error) {
	var migration app.Migration
	res := a.db.WithContext(ctx).
		First(&migration, "name = ?", name)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, res.Error
	}

	return true, nil
}

func (a *AutoMigrate) createMigration(ctx context.Context, name string) error {
	migration := app.Migration{
		Name:   name,
		Status: app.MigrationStatusInProgress,
	}
	res := a.db.WithContext(ctx).
		Create(&migration)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (a *AutoMigrate) updateMigrationStatus(ctx context.Context, name string, status app.MigrationStatus) error {
	currentApp := app.Migration{}
	res := a.db.WithContext(ctx).
		Model(&currentApp).
		Where("name = ?", name).
		Updates(app.Migration{
			Status: status,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to migration app: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("migration not found: %s: %w", name, gorm.ErrRecordNotFound)
	}

	return nil
}

func (a *AutoMigrate) execMigration(ctx context.Context, migration migrations.Migration) error {
	if migration.Disabled {
		return nil
	}

	isApplied, err := a.isMigrationApplied(ctx, migration.Name)
	if err != nil {
		return fmt.Errorf("unable to see if %s was applied", migration.Name)
	}
	if isApplied {
		a.metricsWriter.Incr("migration.count", metrics.ToTags(map[string]string{
			"status": "already_applied",
		}))
		a.l.Debug("migration already applied", zap.String("name", migration.Name))
		return nil
	}

	status := "error"
	statusDescription := ""
	defer func() {
		a.metricsWriter.Event(&statsd.Event{
			Title: "migration",
			Text:  fmt.Sprintf("migration %s", migration.Name),
			Tags: metrics.ToTags(map[string]string{
				"status":             status,
				"status_description": statusDescription,
			}),
		})
		a.metricsWriter.Incr("migration.count", metrics.ToTags(map[string]string{
			"status":             status,
			"status_description": statusDescription}))
	}()

	if err := a.createMigration(ctx, migration.Name); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			a.l.Info("migration already in progress", zap.String("name", migration.Name))
			statusDescription = "already_in_progress"
			return nil
		}

		statusDescription = "db"
		return fmt.Errorf("unable to create migration: %w", err)
	}

	if err := migration.Fn(ctx); err != nil {
		statusDescription = "unable_to_exec"
		if updateErr := a.updateMigrationStatus(ctx, migration.Name, app.MigrationStatusError); updateErr != nil {
			a.l.Info("unable to update migration status", zap.Error(err))
		}
		return err
	}

	if err := a.updateMigrationStatus(ctx, migration.Name, app.MigrationStatusApplied); err != nil {
		a.l.Info("unable to update migration status", zap.Error(err))
		statusDescription = "unable_to_update_migration_status"
	}

	status = "ok"
	statusDescription = "ok"
	return nil
}

func (a *AutoMigrate) execMigrations(ctx context.Context) error {
	migrations := a.migrations.GetAll()

	for _, migration := range migrations {
		if err := a.execMigration(ctx, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}

		a.metricsWriter.Flush()
	}

	return nil
}
