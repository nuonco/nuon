package ch

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type ChModel interface {
	GetTableOptions() (string, bool)
	MigrateDB(tx *gorm.DB) *gorm.DB
}

func AllChModels() []ChModel {
	return []ChModel{
		&app.RunnerHeartBeat{},
		&app.RunnerHealthCheck{},
		&app.OtelLogRecord{},
		&app.OtelTrace{},
		&app.OtelMetricSum{},
		&app.OtelMetricGauge{},
		&app.OtelMetricHistogram{},
		&app.OtelMetricExponentialHistogram{},

		// NOTE(jm): this is a special table used in both ch and postgres
		&app.TableSize{},
	}
}

func AllModels() []interface{} {
	models := AllChModels()
	ifaces := make([]interface{}, len(models))
	for i, model := range models {
		ifaces[i] = model
	}
	return ifaces
}
