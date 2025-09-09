package ch

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func AllModels() []any {
	return []any{
		&app.RunnerHeartBeat{},
		&app.RunnerHealthCheck{},
		&app.OtelLogRecord{},
		&app.OtelTrace{},
		&app.OtelMetricSum{},
		&app.OtelMetricGauge{},
		&app.OtelMetricHistogram{},
		&app.OtelMetricExponentialHistogram{},

		&app.CHTableSize{},

		// noted but not migrated
		// &app.LatestRunnerHeartBeat{},
	}
}
