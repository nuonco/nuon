package service

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const metricRunnerJobPickupAgeMs = "runner_job.pickup_age_ms"

const (
	pickupPathLongPoll = "longpoll"
	pickupPathLegacy   = "legacy"
)

func (s *service) emitRunnerJobPickupAge(jobs []*app.RunnerJob, path string) {
	now := time.Now()
	for _, j := range jobs {
		if j == nil || j.Status != app.RunnerJobStatusAvailable {
			continue
		}
		s.mw.Timing(metricRunnerJobPickupAgeMs, now.Sub(j.CreatedAt), []string{
			"pickup_path:" + path,
		})
	}
}
