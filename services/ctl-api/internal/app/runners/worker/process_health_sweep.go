package worker

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/processhealthcheck"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type ProcessHealthSweepRequest struct{}

// ProcessHealthSweep is the native-scheduling replacement for the per-process
// healthcheck cron emitters. A single Temporal Schedule fires this workflow; it
// lists active/offline processes and enqueues the existing process_healthcheck
// signal onto each process's queue. The signal handler (and its full state
// machine) is unchanged - only the trigger moves from per-process crons to one
// central sweep, so per-owner serialization and dedup are preserved.
//
// @temporal-gen-v2 workflow
func (w *Workflows) ProcessHealthSweep(ctx workflow.Context, _ ProcessHealthSweepRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// Work-layer gate: legacy per-process emitters are authoritative when native
	// scheduling is disabled.
	if !w.cfg.NativeSchedulingEnabled {
		l.Info("native scheduling disabled, process health sweep no-op")
		return nil
	}

	refs, err := activities.AwaitListActiveProcessesForHealthCheck(ctx, &activities.ListActiveProcessesForHealthCheckRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to list active processes")
	}

	l.Info("running process health sweep", zap.Int("process-count", len(refs)))

	for _, ref := range refs {
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   ref.RunnerID,
			OwnerType: "runners",
			QueueName: fmt.Sprintf("runner-process-%s", ref.ProcessID),
			Signal: &processhealthcheck.Signal{
				RunnerID:  ref.RunnerID,
				ProcessID: ref.ProcessID,
			},
		}); err != nil {
			l.Warn("unable to enqueue process health check",
				zap.String("process_id", ref.ProcessID),
				zap.Error(err))
		}
	}

	return nil
}
