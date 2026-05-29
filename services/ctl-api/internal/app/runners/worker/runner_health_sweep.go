package worker

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const runnerHealthSweepChunkSize = 200

type RunnerHealthSweepRequest struct{}

// RunnerHealthSweep is the native-scheduling replacement for the per-runner
// healthcheck cron emitters. A single Temporal Schedule fires this workflow on a
// fixed cadence; it evaluates all active runners against their latest heartbeat in
// batched activities and applies the (rare) status changes via the existing status
// activities. Preserves the signal's behavior (15s threshold, per-tick records,
// warnings) while collapsing the per-runner workflow fleet into one sweep.
//
// @temporal-gen-v2 workflow
func (w *Workflows) RunnerHealthSweep(ctx workflow.Context, _ RunnerHealthSweepRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// Work-layer gate: when native scheduling is disabled the legacy per-runner
	// emitters are authoritative, so the sweep no-ops.
	if !w.cfg.NativeSchedulingEnabled {
		l.Info("native scheduling disabled, runner health sweep no-op")
		return nil
	}

	ids, err := activities.AwaitListActiveRunnerIDsForHealthCheck(ctx, &activities.ListActiveRunnersForHealthCheckRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to list active runners")
	}

	l.Info("running runner health sweep", zap.Int("runner-count", len(ids)))

	for start := 0; start < len(ids); start += runnerHealthSweepChunkSize {
		end := start + runnerHealthSweepChunkSize
		if end > len(ids) {
			end = len(ids)
		}

		resp, err := activities.AwaitEvaluateRunnerHealthChunk(ctx, activities.EvaluateRunnerHealthChunkRequest{
			RunnerIDs: ids[start:end],
		})
		if err != nil {
			return errors.Wrap(err, "unable to evaluate runner health chunk")
		}

		for _, change := range resp.Changes {
			if change.NewStatus != app.RunnerStatusActive {
				l.Warn("runner became unhealthy",
					zap.String("runner_id", change.RunnerID),
					zap.String("new_status", string(change.NewStatus)),
				)
			}

			desc := fmt.Sprintf("status change %s -> %s in health check", change.OldStatus, change.NewStatus)
			if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
				RunnerID:          change.RunnerID,
				Status:            change.NewStatus,
				StatusDescription: desc,
			}); err != nil {
				l.Warn("unable to update runner status", zap.String("runner_id", change.RunnerID), zap.Error(err))
			}

			if err := statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
				RunnerID:          change.RunnerID,
				Status:            change.NewStatus,
				StatusDescription: desc,
			}); err != nil {
				l.Warn("unable to update runner status v2", zap.String("runner_id", change.RunnerID), zap.Error(err))
			}
		}
	}

	return nil
}
