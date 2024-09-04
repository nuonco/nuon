package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql"
)

func (a *AutoMigrate) migratePSQLModels(ctx context.Context) error {
	a.l.Info("running auto migrate for all psql models")

	joinTables := psql.JoinTables()
	for _, joinTable := range joinTables {
		if err := a.psqlDB.WithContext(ctx).SetupJoinTable(joinTable.Model, joinTable.Field, joinTable.JoinTable); err != nil {
			return fmt.Errorf("unable to create join table: %w", err)
		}
	}

	models := psql.AllModels()
	for _, model := range models {
		if err := a.psqlDB.WithContext(ctx).AutoMigrate(model); err != nil {
			return fmt.Errorf("unable to migrate %T: %w", model, err)
		}
	}

	return nil
}
