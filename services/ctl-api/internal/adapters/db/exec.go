package db

import (
	"context"
	"fmt"
)

func (a *AutoMigrate) Execute(ctx context.Context) error {
	if err := a.migrateModels(ctx); err != nil {
		return fmt.Errorf("unable to migrate models: %w", err)
	}

	if err := a.execMigrations(ctx); err != nil {
		return fmt.Errorf("unable to execute migrations: %w", err)
	}

	return nil
}
