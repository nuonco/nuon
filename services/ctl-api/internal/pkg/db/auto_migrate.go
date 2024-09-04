package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/migrations"
)

type Params struct {
	fx.In

	PsqlDB *gorm.DB `name:"psql"`
	CHDB   *gorm.DB `name:"ch"`

	L             *zap.Logger
	Cfg           *internal.Config
	Migrations    *migrations.Migrations
	MetricsWriter metrics.Writer
}

type AutoMigrate struct {
	psqlDB *gorm.DB
	chDB   *gorm.DB

	l             *zap.Logger
	cfg           *internal.Config
	migrations    *migrations.Migrations
	metricsWriter metrics.Writer
}

func NewAutoMigrate(p Params) *AutoMigrate {
	return &AutoMigrate{
		psqlDB:        p.PsqlDB,
		chDB:          p.CHDB,
		l:             p.L,
		cfg:           p.Cfg,
		migrations:    p.Migrations,
		metricsWriter: p.MetricsWriter,
	}
}
