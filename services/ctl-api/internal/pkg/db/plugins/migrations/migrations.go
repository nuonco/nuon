package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
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
				Name:  "indexes",
				Err:   err,
			}
		}
	}

	return nil
}

func (m *Migrator) applyMigration(ctx context.Context, obj any, idx Migration) error {
	mm, ok := m.toMigrationMode(obj)
	if !ok {
		return nil
	}

	for _, migration := range mm.Migrations() {
		if err := m.execMigration(ctx, migration); err != nil {
			return errors.Wrap(err, "migration %s failed: "+migration.Name)
		}

		m.mw.Flush()
	}

	return nil
}

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

	return true, nil
}

func (a *Migrator) createMigration(ctx context.Context, name string) error {
	migration := MigrationModel{
		Name:   name,
		Status: MigrationStatusInProgress,
	}
	res := a.migrationDB.WithContext(ctx).
		Create(&migration)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (a *Migrator) updateMigrationStatus(ctx context.Context, name string, status MigrationStatus) error {
	currentApp := MigrationModel{}
	res := a.migrationDB.WithContext(ctx).
		Model(&currentApp).
		Where("name = ?", name).
		Updates(MigrationModel{
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

	if err := a.createMigration(ctx, migration.Name); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			a.l.Info("migration already in progress", zap.String("name", migration.Name))
			statusDescription = "already_in_progress"
			return nil
		}

		statusDescription = "db"
		return fmt.Errorf("unable to create migration: %w", err)
	}

	if migration.Fn != nil {
		if err := migration.Fn(ctx, a.db); err != nil {
			statusDescription = "unable_to_exec_fn"
			if updateErr := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusError); updateErr != nil {
				a.l.Info("unable to update migration status", zap.Error(err))
			}
			return err
		}
	}

	if migration.SQLFn != nil {
		sql, err := migration.SQLFn(ctx, a.db)
		if err != nil {
			statusDescription = "unable_to_get_sql_sql_fn"
			if updateErr := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusError); updateErr != nil {
				a.l.Info("unable to update migration status", zap.Error(err))
			}
			return err
		}

		res := a.db.WithContext(ctx).Exec(sql)
		if res.Error != nil {
			statusDescription = "unable_to_exec_sql_fn_sql"
			if updateErr := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusError); updateErr != nil {
				a.l.Info("unable to update migration status", zap.Error(err))
			}
			return err
		}

	}

	if migration.SQL != "" {
		res := a.db.WithContext(ctx).Exec(migration.SQL)
		if res.Error != nil {
			statusDescription = "unable_to_exec_sql"
			if updateErr := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusError); updateErr != nil {
				a.l.Info("unable to update migration status", zap.Error(res.Error))
			}
			return errors.Wrap(res.Error, "unable to execute sql")
		}
	}

	if err := a.updateMigrationStatus(ctx, migration.Name, MigrationStatusApplied); err != nil {
		a.l.Info("unable to update migration status", zap.Error(err))
		statusDescription = "unable_to_update_migration_status"
	}

	status = "ok"
	statusDescription = "ok"
	return nil
}
