package helpers

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/processjobsignals"
)

// SignalProcessJob sends the ProcessJob wake-up signal for a single job.
// Used when a known job ID is available (cancellation, execution status
// changes, job created). Fire-and-forget: errors are logged but not
// returned so callers are never broken by a missing/completed workflow.
func (h *Helpers) SignalProcessJob(ctx context.Context, l *zap.Logger, jobID, reason string) {
	err := h.tClient.SignalWorkflowInNamespace(
		ctx,
		signals.TemporalNamespace,
		processjobsignals.WorkflowID(jobID),
		"", // runID "" = latest
		processjobsignals.SignalName,
		processjobsignals.WakeUp{Reason: reason},
	)
	if err != nil {
		// Workflow may have already completed or not yet started — not an error.
		l.Debug("processjob signal skipped",
			zap.String("job_id", jobID),
			zap.String("reason", reason),
			zap.Error(err),
		)
	}
}

// SignalProcessJobsForRunner finds all in-progress jobs for the given runner
// and sends a wake-up signal to each. Used by runner-level writers (heartbeat,
// runner status changes) where the job ID is not directly known.
func (h *Helpers) SignalProcessJobsForRunner(ctx context.Context, l *zap.Logger, runnerID, reason string) {
	var jobs []app.RunnerJob
	h.db.WithContext(ctx).
		Select("id").
		Where(&app.RunnerJob{RunnerID: runnerID}).
		Where("status IN ?", []string{
			string(app.RunnerJobStatusAvailable),
			string(app.RunnerJobStatusQueued),
			string(app.RunnerJobStatusInProgress),
		}).
		Find(&jobs)

	for _, job := range jobs {
		h.SignalProcessJob(ctx, l, job.ID, reason)
	}
}
