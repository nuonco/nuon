package ch

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type ChModel interface {
	GetTableOptions() (string, bool)
	MigrateDB(tx *gorm.DB) *gorm.DB
}

func AllChModels() []ChModel {
	// NOTE(fd): i have to make this manually because i don't understand interface slices enough
	models := make([]ChModel, 8)
	models[0] = app.RunnerHeartBeat{}
	models[1] = app.RunnerHealthCheck{}
	models[2] = app.OtelLogRecord{}
	models[3] = app.OtelTrace{}
	models[4] = app.OtelMetricSum{}
	models[5] = app.OtelMetricGauge{}
	models[6] = app.OtelMetricHistogram{}
	models[7] = app.OtelMetricExponentialHistogram{}
	return models
}

func AllModels() []interface{} {
	models := AllChModels()
	ifaces := make([]interface{}, len(models))
	for i, model := range models {
		ifaces[i] = model
	}
	return ifaces
}
