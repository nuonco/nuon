package migrations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type step struct {
	name string

	objMethod func(context.Context, any) error
}

func (m *Migrator) Exec(ctx context.Context) error {
	methods := []step{
		{
			"join-tables",
			m.applyJoinTables,
		},
		{
			"gorm-migrations",
			m.applyGormMigrations,
		},
		{
			"indexes",
			m.applyIndexes,
		},
		{
			"views",
			m.applyViews,
		},
		{
			"custom-migrations",
			m.applyMigrations,
		},
	}

	for _, method := range methods {
		m.l.Info(fmt.Sprintf("executing %s migration method", method.name),
			zap.String("db_type", m.dbType))

		for _, model := range m.models {
			if err := method.objMethod(ctx, model); err != nil {
				return err
			}
		}
	}

	m.l.Info("applying global migrations",
		zap.String("db_type", m.dbType))
	if err := m.applyGlobalMigrations(ctx); err != nil {
		return errors.Wrap(err, "unable to execute global migrations")
	}

	return nil
}
