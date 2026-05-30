package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Fallback uptime thresholds when config values are not set
const (
	defaultMngUptimeThreshold     = 168 * time.Hour // 1 week
	defaultInstallUptimeThreshold = 8 * time.Hour
	defaultBuildUptimeThreshold   = 8 * time.Hour
)

// CreateProcessQueues creates a queue for the given runner process with a cron health check
// emitter and a scheduled uptime TTL emitter, then enqueues the process_init signal.
func (h *Helpers) CreateProcessQueues(ctx context.Context, runnerID string, process *app.RunnerProcess) (*app.Queue, error) {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerID,
		OwnerType:   "runners",
		Namespace:   signals.TemporalNamespace,
		Name:        fmt.Sprintf("runner-process-%s", process.ID),
		MaxInFlight: 1,
		MaxDepth:    10,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create process queue: %w", err)
	}

	// Cron emitter: process health check once per hour
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         q.ID,
		Name:            fmt.Sprintf("process-%s-health-check", process.ID),
		Description:     "Periodic process health check",
		Mode:            app.QueueEmitterModeCron,
		CronSchedule:    "0 * * * *",
		JitterWindow:    time.Minute * 30,
		SignalType:      "process_healthcheck",
		SignalExpiresIn: 1 * time.Hour,
		SignalTemplate: queuesignal.NewRaw("process_healthcheck", map[string]any{
			"runner_id":  runnerID,
			"process_id": process.ID,
		}),
	}); err != nil {
		return nil, fmt.Errorf("unable to create process health check emitter: %w", err)
	}

	// Scheduled emitter: uptime TTL (from config, with fallback defaults)
	var threshold time.Duration
	switch process.Type {
	case app.RunnerProcessTypeMng:
		threshold = h.cfg.ProcessMngUptimeThreshold
		if threshold == 0 {
			threshold = defaultMngUptimeThreshold
		}
	case app.RunnerProcessTypeBuild:
		threshold = h.cfg.ProcessBuildUptimeThreshold
		if threshold == 0 {
			threshold = defaultBuildUptimeThreshold
		}
	default:
		threshold = h.cfg.ProcessInstallUptimeThreshold
		if threshold == 0 {
			threshold = defaultInstallUptimeThreshold
		}
	}

	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:     q.ID,
		Name:        fmt.Sprintf("process-%s-trigger-shutdown", process.ID),
		Description: "Trigger process shutdown after uptime threshold",
		Mode:        app.QueueEmitterModeFireOnce,
		ScheduledAt: generics.ToPtr(time.Now().Add(threshold)),
		SignalType:  "trigger_shutdown",
		SignalTemplate: queuesignal.NewRaw("trigger_shutdown", map[string]any{
			"runner_id":    runnerID,
			"process_type": string(process.Type),
		}),
	}); err != nil {
		return nil, fmt.Errorf("unable to create trigger shutdown emitter: %w", err)
	}

	// Enqueue the process_init signal to transition process from pending to active
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   process.ID,
		OwnerType: plugins.TableName(h.db, app.RunnerProcess{}),
		Signal: queuesignal.NewRaw("process_init", map[string]any{
			"runner_id":  runnerID,
			"process_id": process.ID,
		}),
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue process init signal: %w", err)
	}

	return q, nil
}
