package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
)

func (a *AutoMigrate) migrateCHModels(ctx context.Context) error {
	a.l.Info("running auto migrate for all clickhouse models")

	models := ch.AllModels()
	for _, model := range models {
		if err := a.chDB.WithContext(ctx).AutoMigrate(model); err != nil {
			return fmt.Errorf("unable to migrate %T: %w", model, err)
		}
	}

	return nil
}
