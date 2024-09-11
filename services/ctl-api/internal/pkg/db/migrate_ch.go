package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
)

func (a *AutoMigrate) migrateCHModels(ctx context.Context) error {
	a.l.Info("running auto migrate for all clickhouse models")

	models := ch.AllChModels()
	for _, chModel := range models {
		// migrate w/ options set
		db := a.chDB.WithContext(ctx)
		db = chModel.MigrateDB(db)
		if err := db.AutoMigrate(chModel); err != nil {
			return fmt.Errorf("unable to migrate %T: %w", chModel, err)
		}
	}

	return nil
}
