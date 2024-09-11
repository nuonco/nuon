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
	models := make([]ChModel, 3)
	models[0] = app.RunnerHeartBeat{}
	models[1] = app.RunnerHealthCheck{}
	models[2] = app.OtelLogRecord{}
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
