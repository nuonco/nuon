package goose

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Goose struct{}

func NewUp(gormDB *gorm.DB, cfg *internal.Config, lc fx.Lifecycle, l *zap.Logger) (*Goose, error) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			l.Info("running migrations")
			db, err := gormDB.DB()
			if err != nil {
				return fmt.Errorf("unable to get database: %w", err)
			}

			if err := goose.Up(db, cfg.DBMigrationsPath); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &Goose{}, nil
}
