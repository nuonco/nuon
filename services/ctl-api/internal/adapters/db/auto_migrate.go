package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db/migrations"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AutoMigrate struct {
	db            *gorm.DB
	l             *zap.Logger
	cfg           *internal.Config
	migrations    *migrations.Migrations
	metricsWriter metrics.Writer
}

func NewAutoMigrate(db *gorm.DB,
	cfg *internal.Config,
	l *zap.Logger,
	migrations *migrations.Migrations,
	metricsWriter metrics.Writer,
	lc fx.Lifecycle,
	shutdowner fx.Shutdowner) *AutoMigrate {
	a := AutoMigrate{
		db:            db,
		l:             l,
		cfg:           cfg,
		migrations:    migrations,
		metricsWriter: metricsWriter,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := a.migrateModels(ctx); err != nil {
				return fmt.Errorf("unable to migrate models: %w", err)
			}

			if err := a.execMigrations(ctx); err != nil {
				return fmt.Errorf("unable to execute migrations: %w", err)
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
