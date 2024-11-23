package ch

import (
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	if err := db.Use(plugins.NewMetricsPlugin(d.MetricsWriter, "ch")); err != nil {
		return errors.Wrap(err, "unable to register metrics plugin")
	}

	if err := db.Use(plugins.NewAfterQueryPlugin()); err != nil {
		return errors.Wrap(err, "unable to register after query plugin")
	}

	if err := db.Use(plugins.NewViewsPlugin(AllModels())); err != nil {
		return errors.Wrap(err, "unable to register views plugin")
	}

	return nil
}
