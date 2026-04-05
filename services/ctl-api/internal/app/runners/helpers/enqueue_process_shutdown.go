package helpers

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

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

// EnqueueProcessShutdown enqueues the process_shutdown signal to the process's
// queue and stops the queue emitters (health check cron, uptime check) so they
// no longer fire for this process.
func (h *Helpers) EnqueueProcessShutdown(ctx context.Context, process *app.RunnerProcess) error {
	q, err := h.queueClient.GetQueueByOwnerAndName(ctx, process.RunnerID, "runners", fmt.Sprintf("runner-process-%s", process.ID))
	if err != nil {
		return fmt.Errorf("unable to find process queue: %w", err)
	}

	// Stop emitters first so health checks stop firing
	if err := h.StopProcessQueue(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to stop process queue emitters: %w", err)
	}

	// Enqueue the process_shutdown signal
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &processShutdownSignalTemplate{
			RunnerID:  process.RunnerID,
			ProcessID: process.ID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue process_shutdown signal: %w", err)
	}

	return nil
}
