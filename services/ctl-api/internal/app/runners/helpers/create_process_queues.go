package helpers

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
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

var processTracer = otel.Tracer("github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers")

func traceProcessOperation(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := processTracer.Start(ctx, name)
	defer span.End()

	if err := fn(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// CreateProcessQueues creates a queue for the given runner process with a
// scheduled uptime TTL emitter, then enqueues the process_init signal. For
// sweep-enabled orgs health checks arrive from the per-org sweep emitter;
// otherwise a legacy per-process cron emitter is created.
func (h *Helpers) CreateProcessQueues(ctx context.Context, runnerID string, process *app.RunnerProcess) (*app.Queue, error) {
	var q *app.Queue
	err := traceProcessOperation(ctx, "runner.process.queue.create", func(ctx context.Context) error {
		var err error
		q, err = h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runnerID,
			OwnerType:   "runners",
			Namespace:   "runners",
			Name:        fmt.Sprintf("runner-process-%s", process.ID),
			MaxInFlight: 1,
			MaxDepth:    10,
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create process queue: %w", err)
	}

	var sweeps bool
	err = traceProcessOperation(ctx, "runner.process.healthcheck_sweeps.evaluate", func(ctx context.Context) error {
		var err error
		sweeps, err = h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, process.OrgID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if !sweeps {
		if err := traceProcessOperation(ctx, "runner.process.healthcheck_emitter.create", func(ctx context.Context) error {
			return h.createProcessHealthcheckEmitter(ctx, q.ID, runnerID, process.ID)
		}); err != nil {
			return nil, err
		}
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

	err = traceProcessOperation(ctx, "runner.process.shutdown_emitter.create", func(ctx context.Context) error {
		_, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
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
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create trigger shutdown emitter: %w", err)
	}

	// Enqueue the process_init signal to transition process from pending to active
	err = traceProcessOperation(ctx, "runner.process.init_signal.enqueue", func(ctx context.Context) error {
		_, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   process.ID,
			OwnerType: plugins.TableName(h.db, app.RunnerProcess{}),
			Signal: queuesignal.NewRaw("process_init", map[string]any{
				"runner_id":  runnerID,
				"process_id": process.ID,
			}),
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to enqueue process init signal: %w", err)
	}

	return q, nil
}

func (h *Helpers) createProcessHealthcheckEmitter(ctx context.Context, queueID, runnerID, processID string) error {
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         queueID,
		Name:            fmt.Sprintf("process-%s-health-check", processID),
		Description:     "Periodic process health check",
		Mode:            app.QueueEmitterModeCron,
		CronSchedule:    ProcessHealthcheckSchedule,
		JitterWindow:    cronutil.MaxJitterWindow,
		SignalType:      "process_healthcheck",
		SignalExpiresIn: runnerHealthcheckSignalExpiry,
		SignalTemplate: queuesignal.NewRaw("process_healthcheck", map[string]any{
			"runner_id":  runnerID,
			"process_id": processID,
		}),
	}); err != nil {
		return fmt.Errorf("unable to create process health check emitter: %w", err)
	}
	return nil
}
