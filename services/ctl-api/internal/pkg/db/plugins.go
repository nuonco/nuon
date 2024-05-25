package db

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	db.Use(plugins.NewMetricsPlugin(d.MetricsWriter))
	db.Use(plugins.NewAfterQueryPlugin())
	db.Use(plugins.NewViewsPlugin(allModels()))

	return nil
}
