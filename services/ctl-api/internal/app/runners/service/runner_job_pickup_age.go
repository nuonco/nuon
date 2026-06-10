package service

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// metricRunnerJobPickupAgeMs records how long a job sat available before a
// runner picked it up, tagged by which pickup path served it. It is emitted on
// both the legacy 5s-poll endpoint and the long-poll/NOTIFY tail endpoint so the
// two can be compared directly: with the long-poll flag on, a freshly-available
// job is handed back in ~ms; on the legacy path it waits out the runner's
// idle-poll interval (up to ~5s).
const metricRunnerJobPickupAgeMs = "runner_job.pickup_age_ms"

const (
	pickupPathLongPoll = "longpoll"
	pickupPathLegacy   = "legacy"
)

// emitRunnerJobPickupAge emits the pickup-age timing for each available job
// being handed back to a runner. The baseline is CreatedAt: interactive jobs
// (actions, sandbox, notebook cells) are born `available`, so CreatedAt is the
// instant the job became pickable and the timing is exact pickup latency. Jobs
// that queue before becoming available will read slightly high (the wait
// includes queue time).
//
// Caveat: the legacy endpoint can return the same still-available job on
// successive polls until the runner claims it, so a backlogged job may be
// sampled more than once — this only ever inflates the legacy distribution,
// never the long-poll one (which returns a job once on its NOTIFY wake).
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
