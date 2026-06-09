package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// wakeProcessJobWorkflow best-effort signals the process_job handler workflow
// for runnerJobID so its poll loop reacts to a state change (the runner
// reserving the job, or reporting a terminal status) on the request we already
// receive, instead of waiting out its next poll tick. The caller supplies the
// edge-specific signal name (see processjob.PickupSignalName /
// processjob.TerminalSignalName) so the pickup and finalize loops never consume
// each other's wake.
//
// It is intentionally fire-and-forget: the workflow's poll timer is the
// correctness guarantee, so a missing signal row, an already-finished
// workflow, or a transient Temporal error just falls back to polling. We never
// fail the runner's request because of it.
func (s *service) wakeProcessJobWorkflow(ctx context.Context, runnerJobID, signalName string) {
	if s.temporalClient == nil {
		return
	}

	// The handler workflow ref is persisted on the process_job QueueSignal row,
	// keyed by the job (owner). There is at most one such signal per job.
	var qs app.QueueSignal
	err := s.db.WithContext(ctx).
		Where(app.QueueSignal{
			// matches the OwnerType set when the process_job signal is
			// enqueued (see pkg/workflows/job queueJob).
			OwnerType: "runner_jobs",
			OwnerID:   runnerJobID,
		}).
		Order("created_at desc").
		First(&qs).Error
	if err != nil {
		// Legacy poll-dispatch path, or the signal hasn't been enqueued yet —
		// nothing to wake. The workflow (if any) will detect the change on its
		// next poll.
		return
	}
	if qs.Workflow.ID == "" {
		return
	}

	if err := s.temporalClient.SignalWorkflowInNamespace(
		ctx,
		qs.Workflow.Namespace,
		qs.Workflow.ID,
		"", // empty RunID = latest run
		signalName,
		nil,
	); err != nil {
		s.l.Warn("unable to wake process_job workflow",
			zap.String("runner_job.id", runnerJobID),
			zap.String("workflow.id", qs.Workflow.ID),
			zap.String("signal", signalName),
			zap.Error(err))
	}
}
