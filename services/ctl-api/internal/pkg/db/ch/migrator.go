package ch

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	chmigrations "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ChParams struct {
	fx.In

	Migrations   *chmigrations.Migrations
	MigrationsDB *gorm.DB `name:"psql"`
	DB           *gorm.DB `name:"ch"`

	L             *zap.Logger
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
}

func NewCHMigrator(p ChParams, lc fx.Lifecycle) *migrations.Migrator {
	opts := migrations.NewOpts()
	opts.CreateViewSQLTmpl = "CREATE OR REPLACE VIEW %s ON CLUSTER simple AS %s"

	return migrations.New(migrations.Params{
		Opts:         opts,
		Migrations:   p.Migrations.All(),
		Models:       AllModels(),
		MigrationsDB: p.MigrationsDB,
		DB:           p.DB,
		DBType:       "ch",
		L:            p.L,
		Cfg:          p.Cfg,
		MW:           p.MetricsWriter,
		TableOpts: map[string]string{
			"gorm:table_cluster_options": "on cluster simple",
		},
	})
}
