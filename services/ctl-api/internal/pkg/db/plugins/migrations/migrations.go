package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type Migration struct {
	Name      string
	Disabled  bool
	AlwaysRun bool

	Fn    func(context.Context, *gorm.DB) error
	SQL   string
	SQLFn func(context.Context, *gorm.DB) (string, error)
}

type migrationModel interface {
	Migrations() []Migration
}

func (m *Migrator) toMigrationMode(obj any) (migrationModel, bool) {
	jtm, ok := obj.(migrationModel)
	return jtm, ok
}

func (m *Migrator) applyMigrations(ctx context.Context, obj any) error {
	mm, ok := m.toMigrationMode(obj)
	if !ok {
		return nil
	}

	for _, idx := range mm.Migrations() {
		if err := m.applyMigration(ctx, obj, idx); err != nil {
			return MigrationErr{
				Model: plugins.TableName(m.db, obj),
				Name:  idx.Name,
				Err:   err,
			}
		}
	}

	return nil
}

func (m *Migrator) applyMigration(ctx context.Context, _ any, idx Migration) error {
	if err := m.execMigration(ctx, idx); err != nil {
		return errors.Wrap(err, "migration failed: "+idx.Name)
	}

	m.mw.Flush()

	return nil
}

// an in_progress migration older than this is assumed dead and gets reclaimed
const abandonedMigrationTimeout = time.Hour

func (a *Migrator) isMigrationApplied(ctx context.Context, name string) (bool, error) {
	var migration MigrationModel
	res := a.migrationDB.WithContext(ctx).
		First(&migration, "name = ?", name)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, res.Error
	}

	// only applied counts. an error row used to read as applied and never retried.
	return migration.Status == MigrationStatusApplied, nil
}

// claimMigration marks a migration in progress and reports whether we won the claim.
// errored and abandoned rows get reclaimed so they retry.
func (a *Migrator) claimMigration(ctx context.Context, name string) (bool, error) {
	now := time.Now()

	res := a.migrationDB.WithContext(ctx).
		Model(&MigrationModel{}).
		Where("name = ?", name).
		Where("status = ? OR (status = ? AND updated_at < ?)",
			MigrationStatusError,
			MigrationStatusInProgress, now.Add(-abandonedMigrationTimeout)).
		Updates(map[string]any{
			"status":     MigrationStatusInProgress,
			"updated_at": now,
		})
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected > 0 {
		return true, nil
	}

	// nothing to reclaim, first attempt
	migration := MigrationModel{
		Name:   name,
		Status: MigrationStatusInProgress,
	}
	if err := a.migrationDB.WithContext(ctx).Create(&migration).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// another run got there first
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// maxMigrationErrLen keeps a runaway error string from bloating the row
const maxMigrationErrLen = 2000

func (a *Migrator) updateMigrationStatus(ctx context.Context, name string, status MigrationStatus, cause error) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
		"error":      nil,
	}
	if cause != nil {
		msg := cause.Error()
		if len(msg) > maxMigrationErrLen {
			msg = msg[:maxMigrationErrLen]
		}
		updates["error"] = msg
	}

	res := a.migrationDB.WithContext(ctx).
		Model(&MigrationModel{}).
		Where("name = ?", name).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("unable to migration app: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("migration not found: %s: %w", name, gorm.ErrRecordNotFound)
	}

	return nil
}

func (a *Migrator) execMigration(ctx context.Context, migration Migration) error {
	if migration.Disabled {
		return nil
	}

	if !migration.AlwaysRun {
		isApplied, err := a.isMigrationApplied(ctx, migration.Name)
		if err != nil {
			return fmt.Errorf("unable to see if %s was applied", migration.Name)
		}
		if isApplied {
			a.mw.Incr("migration.count", metrics.ToTags(map[string]string{
				"db_type": a.dbType,
				"status":  "already_applied",
			}))
			a.l.Debug("migration already applied", zap.String("name", migration.Name))
			return nil
		}
	} else {
		a.l.Info("running migration without checking because of `AlwaysRun`")
	}

	status := "error"
	statusDescription := ""
	defer func() {
		a.mw.Event(&statsd.Event{
			Title: "migration",
			Text:  fmt.Sprintf("migration %s", migration.Name),
			Tags: metrics.ToTags(map[string]string{
				"db_type":            a.dbType,
				"status":             status,
				"status_description": statusDescription,
			}),
		})
		a.mw.Incr("migration.count", metrics.ToTags(map[string]string{
			"db_type":            a.dbType,
			"status":             status,
			"status_description": statusDescription,
		}))
	}()

	if migration.AlwaysRun {
		// Note(jm): this is so we can re-run migrations, but not on every single deploy (to prevent killing the
		// database in a case where we are flapping)
		ts := time.Now().Round(time.Hour * 1)
		if a.cfg.Env == config.Development {
			ts = time.Now()
		}

		migration.Name = fmt.Sprintf("%s-%d", migration.Name, ts.Unix())
	}

	claimed, err := a.claimMigration(ctx, migration.Name)
	if err != nil {
		statusDescription = "db"
		return fmt.Errorf("unable to claim migration: %w", err)
	}
	if !claimed {
		a.l.Info("migration already in progress", zap.String("name", migration.Name))
		statusDescription = "already_in_progress"
		return nil
	}

	fail := func(cause error, desc string) error {
		statusDescription = desc
		a.l.Error("migration failed",
			zap.String("name", migration.Name),
			zap.String("db_type", a.dbType),
			zap.Error(cause))

		if updateErr := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusError, cause); updateErr != nil {
			a.l.Warn("unable to update migration status", zap.Error(updateErr))
		}

		return cause
	}

	if migration.Fn != nil {
		if err := migration.Fn(ctx, a.db); err != nil {
			return fail(err, "unable_to_exec_fn")
		}
	}

	if migration.SQLFn != nil {
		sql, err := migration.SQLFn(ctx, a.db)
		if err != nil {
			return fail(err, "unable_to_get_sql_sql_fn")
		}

		// this used to return the nil err from SQLFn above, reporting a failed migration
		// as a success
		if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
			return fail(res.Error, "unable_to_exec_sql_fn_sql")
		}
	}

	if migration.SQL != "" {
		if res := a.db.WithContext(ctx).Exec(migration.SQL); res.Error != nil {
			return fail(errors.Wrap(res.Error, "unable to execute sql"), "unable_to_exec_sql")
		}
	}

	if err := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusApplied, nil); err != nil {
		a.l.Info("unable to update migration status", zap.Error(err))
		statusDescription = "unable_to_update_migration_status"
	}

	status = "ok"
	statusDescription = "ok"
	return nil
}
