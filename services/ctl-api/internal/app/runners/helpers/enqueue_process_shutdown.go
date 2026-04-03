package helpers

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// processShutdownSignalType mirrors the constant in runners/signals/v2/processshutdown
// (defined here to avoid an import cycle through worker/activities).
const processShutdownSignalType queuesignal.SignalType = "process_shutdown"

type processShutdownSignalTemplate struct {
	RunnerID     string `json:"runner_id"`
	ProcessID    string `json:"process_id"`
	ShutdownType string `json:"shutdown_type"`
}

func (s *processShutdownSignalTemplate) Type() queuesignal.SignalType {
	return processShutdownSignalType
}
func (s *processShutdownSignalTemplate) Validate(_ workflow.Context) error { return nil }
func (s *processShutdownSignalTemplate) Execute(_ workflow.Context) error  { return nil }

// EnqueueProcessShutdown enqueues a process_shutdown signal to the process's v2 queue,
// then fully terminates the queue (stops and deletes emitters, stops workflow, soft-deletes queue).
func (h *Helpers) EnqueueProcessShutdown(ctx context.Context, runnerID, processID string, shutdownType app.RunnerProcessShutdownType) error {
	queueName := fmt.Sprintf("runner-process-%s", processID)

	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: "runners",
			Name:      queueName,
		}).First(&q); res.Error != nil {
		return fmt.Errorf("unable to find process queue %s: %w", queueName, res.Error)
	}

	// Enqueue the shutdown signal to the process queue before terminating
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &processShutdownSignalTemplate{
			RunnerID:     runnerID,
			ProcessID:    processID,
			ShutdownType: string(shutdownType),
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue process shutdown signal: %w", err)
	}

	// Terminate the queue: stop and delete emitters, stop workflow, soft-delete queue
	if err := h.queueClient.Terminate(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to terminate process queue: %w", err)
	}

	return nil
}
