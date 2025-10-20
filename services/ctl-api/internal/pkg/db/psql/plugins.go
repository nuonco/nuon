package psql

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/afterquery"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/pagination"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

func (d *database) registerPlugins(db *gorm.DB) error {
	db.Use(metrics.NewMetricsPlugin(d.MetricsWriter, "psql", &d.Logger))
	db.Use(afterquery.NewAfterQueryPlugin())
	db.Use(views.NewViewsPlugin(AllModels()))
	db.Use(pagination.NewOffsetPaginationPlugin())
	db.Use(patcher.NewPatcherPlugin())

	return nil
}
