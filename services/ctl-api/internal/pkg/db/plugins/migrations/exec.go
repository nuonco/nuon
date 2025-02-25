package migrations

import (
	"context"

	"github.com/pkg/errors"
)

func (m *Migrator) Exec(ctx context.Context) error {
	methods := []func(context.Context, any) error{
		// join tables must be declared _before_ the auto-migrate due to constraints
		m.applyJoinTables,
		m.applyGormMigrations,

		// custom migrations
		m.applyIndexes,
		m.applyViews,
		m.applyMigrations,
	}
	for _, method := range methods {
		for _, model := range m.models {
			if err := method(ctx, model); err != nil {
				return err
			}
		}
	}

	if err := m.applyGlobalMigrations(ctx); err != nil {
		return errors.Wrap(err, "unable to execute global migrations")
	}

	return nil
}
