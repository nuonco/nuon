package psql

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	psqlmigrations "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql/migrations"
)

type PSQLParams struct {
	fx.In

	Migrations   *psqlmigrations.Migrations
	MigrationsDB *gorm.DB `name:"psql"`
	DB           *gorm.DB `name:"psql"`

	L             *zap.Logger
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
}

func NewPSQLMigrator(p PSQLParams, lc fx.Lifecycle) *migrations.Migrator {
	opts := migrations.NewOpts()
	mig := migrations.New(migrations.Params{
		Models:       AllModels(),
		Migrations:   p.Migrations.All(),
		MigrationsDB: p.MigrationsDB,
		DB:           p.DB,
		L:            p.L,
		Cfg:          p.Cfg,
		MW:           p.MetricsWriter,
		Opts:         opts,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return mig.Exec(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return mig
}
