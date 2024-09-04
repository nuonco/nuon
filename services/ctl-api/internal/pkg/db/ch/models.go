package ch

import "github.com/powertoolsdev/mono/services/ctl-api/internal/app"

func AllModels() []interface{} {
	return []interface{}{
		app.RunnerHeartBeat{},
		app.RunnerHealthCheck{},

		// logs, metrics, events
		// app.OtelLogRecord{},
		// app.OtelMetricSum{},
		// app.OtelMetricGauge{},
		// app.OtelMetricHistogram{},
		// app.OtelTrace{},
		// app.Event{},

		// internal events
		// app.InstallEvent{},
	}
}
