package helpers

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
)

// UpdateProcessJob fires a workflow update at ProcessJob for the given job
// ID. Fire-and-forget: failures are logged at debug level because the
// ProcessJob workflow may not be running (job not yet queued, or already
// terminal), and that is a normal case.
//
// updateName should be one of worker.UpdateNameXxx constants. payload is
// advisory; ProcessJob re-reads authoritative state from the DB on wake.
func (h *Helpers) UpdateProcessJob(ctx context.Context, jobID, updateName string, payload any) {
	if jobID == "" || updateName == "" {
		return
	}
	workflowID := fmt.Sprintf("processjob-%s", jobID)
	_, err := h.tClient.UpdateWorkflowInNamespace(ctx, signals.TemporalNamespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   workflowID,
		UpdateName:   updateName,
		WaitForStage: tclient.WorkflowUpdateStageAccepted,
		Args:         []any{payload},
	})
	if err != nil {
		h.l.Debug("failed to update process job workflow",
			zap.String("job-id", jobID),
			zap.String("update-name", updateName),
			zap.Error(err),
		)
	}
}

// UpdateProcessJobsForRunner fans out an update to every ProcessJob
// workflow currently tracking an active job for the given runner. Used for
// runner-scoped events (runner status change, runner restart) where the
// single event affects multiple in-flight jobs.
func (h *Helpers) UpdateProcessJobsForRunner(ctx context.Context, runnerID, updateName string, payload any) {
	if runnerID == "" || updateName == "" {
		return
	}
	var jobs []app.RunnerJob
	if err := h.db.WithContext(ctx).
		Select("id").
		Where(&app.RunnerJob{RunnerID: runnerID}).
		Where("status IN ?", []string{
			string(app.RunnerJobStatusAvailable),
			string(app.RunnerJobStatusQueued),
			string(app.RunnerJobStatusInProgress),
		}).
		Find(&jobs).Error; err != nil {
		h.l.Warn("failed to query active jobs for runner update fan-out",
			zap.String("runner-id", runnerID),
			zap.Error(err),
		)
		return
	}
	for _, j := range jobs {
		h.UpdateProcessJob(ctx, j.ID, updateName, payload)
	}
}
