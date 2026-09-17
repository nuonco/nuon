package worker

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/bulkcancel"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	bulkCancelPerRun       = 200
	bulkCancelAttempts     = 3
	bulkCancelTimeout      = 2 * time.Minute
	bulkCancelMaxFailedIDs = 50
)

// BulkCancelWorkflows cancels workflows one at a time, counting failures rather
// than aborting so one bad row does not strand the rest of the backlog.
func (w *Workflows) BulkCancelWorkflows(ctx workflow.Context, req bulkcancel.Request) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if req.Total == 0 {
		req.Total = len(req.Pending)
	}

	progress := bulkcancel.Progress{
		Total:     req.Total,
		Done:      req.Cancelled + req.Skipped + req.Failed,
		Cancelled: req.Cancelled,
		Skipped:   req.Skipped,
		Failed:    req.Failed,
		FailedIDs: req.FailedIDs,
	}
	if err := workflow.SetQueryHandler(ctx, bulkcancel.ProgressQueryType, func() (bulkcancel.Progress, error) {
		return progress, nil
	}); err != nil {
		return errors.Wrap(err, "unable to register bulk cancel progress query handler")
	}

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: bulkCancelTimeout,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: bulkCancelAttempts},
	})

	processedThisRun := 0
	for len(req.Pending) > 0 && processedThisRun < bulkCancelPerRun {
		workflowID := req.Pending[0]
		req.Pending = req.Pending[1:]
		processedThisRun++

		progress.CurrentWorkflowID = workflowID
		resp, err := activities.AwaitCancelWorkflowByWorkflowID(actCtx, workflowID)
		if err != nil {
			req.Failed++
			if len(req.FailedIDs) < bulkCancelMaxFailedIDs {
				req.FailedIDs = append(req.FailedIDs, workflowID)
			}
			l.Warn("unable to cancel workflow", zap.String("workflow_id", workflowID), zap.Error(err))
		} else if resp.Outcome == activities.CancelWorkflowOutcomeSkipped {
			req.Skipped++
		} else {
			req.Cancelled++
		}

		progress.Cancelled = req.Cancelled
		progress.Skipped = req.Skipped
		progress.Failed = req.Failed
		progress.FailedIDs = req.FailedIDs
		progress.Done = req.Cancelled + req.Skipped + req.Failed
	}

	if len(req.Pending) > 0 {
		l.Info("continuing bulk workflow cancellation",
			zap.Int("done", progress.Done),
			zap.Int("total", req.Total))
		return workflow.NewContinueAsNewError(ctx, bulkcancel.WorkflowName, req)
	}

	progress.CurrentWorkflowID = ""
	progress.Finished = true
	l.Info("bulk workflow cancellation complete",
		zap.String("reason", req.Reason),
		zap.Int("cancelled", req.Cancelled),
		zap.Int("skipped", req.Skipped),
		zap.Int("failed", req.Failed))
	return nil
}
