package helpers

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// processHealthcheckSignalType mirrors the constant in runners/signals/v2/processhealthcheck
// (defined here to avoid an import cycle through worker/activities).
const processHealthcheckSignalType queuesignal.SignalType = "process_healthcheck"

type processHealthcheckSignalTemplate struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

func (s *processHealthcheckSignalTemplate) Type() queuesignal.SignalType {
	return processHealthcheckSignalType
}
func (s *processHealthcheckSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *processHealthcheckSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// triggerShutdownSignalType mirrors the constant in runners/signals/v2/triggershutdown
const triggerShutdownSignalType queuesignal.SignalType = "trigger_shutdown"

type triggerShutdownSignalTemplate struct {
	RunnerID    string `json:"runner_id"`
	ProcessType string `json:"process_type"`
}

func (s *triggerShutdownSignalTemplate) Type() queuesignal.SignalType {
	return triggerShutdownSignalType
}
func (s *triggerShutdownSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *triggerShutdownSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// processStartedSignalType mirrors the constant in runners/signals/v2/processstarted
const processStartedSignalType queuesignal.SignalType = "process_started"

type processStartedSignalTemplate struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

func (s *processStartedSignalTemplate) Type() queuesignal.SignalType {
	return processStartedSignalType
}
func (s *processStartedSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *processStartedSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// processCreatedSignalType mirrors the constant in runners/signals/v2/processcreated
const processCreatedSignalType queuesignal.SignalType = "process_created"

type processCreatedSignalTemplate struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

func (s *processCreatedSignalTemplate) Type() queuesignal.SignalType {
	return processCreatedSignalType
}
func (s *processCreatedSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *processCreatedSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// onRunnerProcessSignalType mirrors the constant in runners/signals/v2/onrunnerprocess
const onRunnerProcessSignalType queuesignal.SignalType = "on_runner_process"

type onRunnerProcessSignalTemplate struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`
}

func (s *onRunnerProcessSignalTemplate) Type() queuesignal.SignalType {
	return onRunnerProcessSignalType
}
func (s *onRunnerProcessSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *onRunnerProcessSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// Fallback uptime thresholds when config values are not set
const (
	defaultMngUptimeThreshold     = 168 * time.Hour // 1 week
	defaultInstallUptimeThreshold = 8 * time.Hour
	defaultBuildUptimeThreshold   = 8 * time.Hour
)

// CreateProcessQueues creates a queue for the given runner process with a cron health check
// emitter and a scheduled uptime TTL emitter, then enqueues the process_created signal.
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

	// Cron emitter: process health check every minute
	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        q.ID,
		Name:           fmt.Sprintf("process-%s-health-check", process.ID),
		Description:    "Periodic process health check",
		Mode:           app.QueueEmitterModeCron,
		CronSchedule:   "* * * * *",
		SignalType:     processHealthcheckSignalType,
		SignalTemplate: &processHealthcheckSignalTemplate{RunnerID: runnerID, ProcessID: process.ID},
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
		SignalType:  triggerShutdownSignalType,
		SignalTemplate: &triggerShutdownSignalTemplate{
			RunnerID:    runnerID,
			ProcessType: string(process.Type),
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to create trigger shutdown emitter: %w", err)
	}

	// Enqueue the process_started signal to transition process from pending to active
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &processStartedSignalTemplate{
			RunnerID:  runnerID,
			ProcessID: process.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue process started signal: %w", err)
	}

	// Enqueue the process_created signal to update runner status
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &processCreatedSignalTemplate{
			RunnerID:  runnerID,
			ProcessID: process.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue process created signal: %w", err)
	}

	// Enqueue the on_runner_process signal to notify of the new process
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &onRunnerProcessSignalTemplate{
			RunnerID:  runnerID,
			ProcessID: process.ID,
		},
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue on_runner_process signal: %w", err)
	}

	return q, nil
}
