package db

import (
	"context"
	"fmt"
)

func (a *AutoMigrate) Execute(ctx context.Context) error {
	// NOTE(jm): this is a temporary change, and needs to be reverted
	if err := a.execMigrations(ctx); err != nil {
		return fmt.Errorf("unable to execute migrations: %w", err)
	}

	if err := a.migrateModels(ctx); err != nil {
		return fmt.Errorf("unable to migrate models: %w", err)
	}

	return nil
}
