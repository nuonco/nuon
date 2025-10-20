package ch

import (
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/afterquery"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	if err := db.Use(metrics.NewMetricsPlugin(d.MetricsWriter, "ch", &d.Logger)); err != nil {
		return errors.Wrap(err, "unable to register metrics plugin")
	}

	if err := db.Use(afterquery.NewAfterQueryPlugin()); err != nil {
		return errors.Wrap(err, "unable to register after query plugin")
	}

	if err := db.Use(views.NewViewsPlugin(AllModels())); err != nil {
		return errors.Wrap(err, "unable to register views plugin")
	}

	return nil
}
