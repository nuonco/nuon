package psql

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/afterquery"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	db.Use(metrics.NewMetricsPlugin(d.MetricsWriter, "psql"))
	db.Use(afterquery.NewAfterQueryPlugin())
	db.Use(views.NewViewsPlugin(AllModels()))

	return nil
}
