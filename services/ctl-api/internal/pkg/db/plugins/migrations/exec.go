package migrations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type step struct {
	name string

	objMethod func(*Migrator, context.Context, any) error
}

func (m *Migrator) Exec(ctx context.Context) error {
	if err := m.migrationDB.WithContext(ctx).AutoMigrate(&MigrationModel{}); err != nil {
		return errors.Wrap(err, "unable to ensure migrations table exists")
	}

	methods := []step{
		{
			"join-tables",
			(*Migrator).applyJoinTables,
		},
		{
			"gorm-migrations",
			(*Migrator).applyGormMigrations,
		},
		{
			"indexes",
			(*Migrator).applyIndexes,
		},
		{
			"views",
			(*Migrator).applyViews,
		},
		{
			"custom-migrations",
			(*Migrator).applyMigrations,
		},
	}

	for _, method := range methods {
		m.l.Info(fmt.Sprintf("executing %s migration method", method.name),
			zap.String("db_type", m.dbType))

		apply := func(m *Migrator) error {
			for _, model := range m.models {
				if err := method.objMethod(m, ctx, model); err != nil {
					tableName := plugins.TableName(m.db, model)

					m.l.Error("unable to migrate model",
						zap.Error(err),
						zap.String("model", tableName))

					return err
				}
			}
			return nil
		}
		if method.name == "views" && m.db.Dialector.Name() == "postgres" {
			if err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				scoped := *m
				scoped.db = tx
				if err := scoped.dropViews(ctx); err != nil {
					return errors.Wrap(err, "unable to drop registered views")
				}
				return apply(&scoped)
			}); err != nil {
				return err
			}
		} else if err := apply(m); err != nil {
			return err
		}
	}

	m.l.Info("applying global migrations",
		zap.String("db_type", m.dbType))
	if err := m.applyGlobalMigrations(ctx); err != nil {
		// the per-model phases above log their own failures; without this a failed global
		// migration produced no log line at all
		m.l.Error("unable to apply global migrations",
			zap.Error(err),
			zap.String("db_type", m.dbType))

		return errors.Wrap(err, "unable to execute global migrations")
	}

	return nil
}
