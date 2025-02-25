package migrations

import (
	"context"
)

func (m *Migrator) applyGlobalMigrations(ctx context.Context) error {
	for _, mig := range m.globalMigrations {
		if err := m.execMigration(ctx, mig); err != nil {
			return MigrationErr{
				Model: "global-migrations",
				Name:  "indexes",
				Err:   err,
			}
		}
	}

	return nil
}
