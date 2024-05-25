package db

import (
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/migrations"
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
	metricsWriter metrics.Writer) *AutoMigrate {
	a := AutoMigrate{
		db:            db,
		l:             l,
		cfg:           cfg,
		migrations:    migrations,
		metricsWriter: metricsWriter,
	}

	return &a
}
