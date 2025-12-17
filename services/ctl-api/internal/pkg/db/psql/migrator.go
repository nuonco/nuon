package psql

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
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
	return migrations.New(migrations.Params{
		Models:       AllModels(),
		Migrations:   p.Migrations.All(),
		MigrationsDB: p.MigrationsDB,
		DB:           p.DB,
		DBType:       "psql",
		L:            p.L,
		Cfg:          p.Cfg,
		MW:           p.MetricsWriter,
		Opts:         opts,
	})
}
