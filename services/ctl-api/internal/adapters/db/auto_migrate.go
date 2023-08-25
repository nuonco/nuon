package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AutoMigrate struct {
	db  *gorm.DB
	l   *zap.Logger
	cfg *internal.Config
}

func NewAutoMigrate(db *gorm.DB, cfg *internal.Config, l *zap.Logger, lc fx.Lifecycle, shutdowner fx.Shutdowner) *AutoMigrate {
	a := AutoMigrate{
		db:  db,
		l:   l,
		cfg: cfg,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := a.migrateModels(ctx); err != nil {
				return fmt.Errorf("unable to migrate models: %w", err)
			}

			if err := a.seedModels(ctx); err != nil {
				return fmt.Errorf("unable to seed models: %w", err)
			}

			if err := shutdowner.Shutdown(); err != nil {
				return fmt.Errorf("unable to shut down: %w", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &a
}
