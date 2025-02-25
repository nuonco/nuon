package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

type tableOptsModel interface {
	GetTableOptions() string
}

func (m *Migrator) applyGormMigrations(ctx context.Context, obj any) error {
	db := m.db

	if tom, ok := obj.(tableOptsModel); ok {
		opts := tom.GetTableOptions()
		db = db.Set("gorm:table_options", opts)
	}

	for k, v := range m.tableOpts {
		db = db.Set(k, v)
	}

	if err := db.
		WithContext(ctx).
		AutoMigrate(obj); err != nil {
		return MigrationErr{
			Model: plugins.TableName(db, obj),
			Name:  "gorm-auto",
			Err:   err,
		}
	}
	return nil
}
