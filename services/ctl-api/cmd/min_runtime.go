package cmd

import (
	"os"
	"time"

	"go.uber.org/zap"
)

const minJobRuntime = 60 * time.Second

// NOTE(fd): in prod and stage, we want the job container to persist for at least 60s to ensure
// the datadog agent picks up on its existence. We don't want this job to take longer than necessary
// though so we calculate it's runtime so we only sleep for as long as necessary to reach the 60s threshold.
func ensureMinRuntime(start time.Time) {
	if os.Getenv("ENV") != "prod" && os.Getenv("ENV") != "stage" {
		return
	}

	runTime := time.Since(start)
	if runTime >= minJobRuntime {
		return
	}

	sleepFor := minJobRuntime - runTime
	zap.L().Info("sleeping to ensure data dog metrics are flushed", zap.Duration("duration", sleepFor))
	time.Sleep(sleepFor)
}
