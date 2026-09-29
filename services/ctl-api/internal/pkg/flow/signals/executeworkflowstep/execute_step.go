package executeworkflowstep

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (s *Signal) Execute(ctx workflow.Context) (err error) {
	defer func() { s.finished = true }()

	ctx = cctx.SetWorkflowTelemetryWorkflowContext(ctx, s.workflowTelemetry())
	workflowType := cctx.WorkflowTelemetryFromContext(ctx).WorkflowType

	if s.mw != nil && s.v != nil {
		tmw, metricsErr := tmetrics.New(s.v, tmetrics.WithMetricsWriter(s.mw))
		if metricsErr == nil {
			s.tmw = tmw
		}
	}

	start := workflow.Now(ctx)
	var stepName, executionType string
	defer func() {
		if s.mw == nil {
			return
		}
		tags := metrics.ToTags(map[string]string{
			"workflow_type":  workflowType,
			"owner_type":     s.OwnerType,
			"step_name":      stepName,
			"execution_type": executionType,
		})

		s.mw.Timing("workflow_step.latency", workflow.Now(ctx).Sub(start), tags)
		s.mw.Incr("workflow_step.executed", tags)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow logger")
	}

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}
	stepName = step.Name
	executionType = string(step.ExecutionType)

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}

	if s.ResumeApproval {
		return s.resumeApproval(ctx, l, step, flw)
	}

	if step.Status.Status != app.StatusPending && step.Status.Status != app.StatusNotAttempted && step.Status.Status != app.StatusQueued {
		l.Debug("step not in executable state, exiting",
			zap.String("step_id", step.ID),
			zap.String("step_status", string(step.Status.Status)),
			zap.String("workflow_id", flw.ID))
		return nil
	}

	defer func() {
		activityCtx := ctx
		if ctx.Err() != nil {
			var cancel workflow.CancelFunc
			activityCtx, cancel = workflow.NewDisconnectedContext(ctx)
			defer cancel()
		}
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(activityCtx, step.ID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + step.Name,
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step")
	}

	stepErr := s.executeInnerSignal(ctx, step)
	if stepErr != nil {
		if ctx.Err() != nil {
			if !s.canceled {
				dctx, dcancel := workflow.NewDisconnectedContext(ctx)
				defer dcancel()
				if err := setResultDirective(dctx, s.StepID, DirectiveStop); err != nil {
					l.Error("unable to set fallback stop directive", zap.Error(err))
				}
				_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(dctx, statusactivities.UpdateStatusRequest{
					ID: s.StepID,
					Status: app.CompositeStatus{
						Status:                 app.StatusError,
						StatusHumanDescription: "step interrupted: context cancelled",
						Metadata: map[string]any{
							"reason": stepErr.Error(),
						},
					},
				})
			}
			return nil
		}
		if callback.IsCancelled(stepErr) {
			return s.handleStepCancelled(ctx, l)
		}
		return s.handleStepError(ctx, l, step, flw, stepErr)
	}

	if s.canceled {
		return nil
	}

	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	if step.ExecutionType != app.WorkflowStepExecutionTypeApproval {
		l.Debug("step type non approval, step successful",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		// why: A signal that skipped its own work marks the step skipped before
		// returning. Overwriting that with success would report work as done
		// that never ran, so leave an already-skipped status alone.
		if !isSkippedStatus(step.Status.Status) {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status: app.StatusSuccess,
				},
			}); err != nil {
				return errors.Wrap(err, "unable to mark step as success")
			}
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "finished executing step " + step.Name,
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update flow status after step")
		}

		return nil
	}

	if s.canceled {
		return nil
	}

	return s.processPlan(ctx, step, flw)
}

func (s *Signal) executeInnerSignal(ctx workflow.Context, step *app.WorkflowStep) error {
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepStartedAtByID(ctx, step.ID); err != nil {
		return err
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}); err != nil {
		return err
	}

	if step.ExecutionType == app.WorkflowStepExecutionTypeSkipped {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	if step.QueueSignal == nil {
		return nil
	}

	sig := step.QueueSignal.Signal
	if sig == nil {
		return nil
	}

	signal.ApplyStepContext(sig, step.ID, s.WorkflowID)

	signal.ApplyRetryCount(sig, step.RetryIndex, step.GroupRetryIdx)

	logger := workflow.GetLogger(ctx)
	logger.Info("enqueuing signal to target queue",
		"step_name", step.Name,
		"step_id", step.ID,
		"owner_id", s.OwnerID,
		"owner_type", s.OwnerType,
		"target_queue", s.TargetQueueName,
	)

	cb := callback.New(ctx, step.ID)
	dedupeKey := fmt.Sprintf("workflow-step:%s:retry:%d:group-retry:%d", step.ID, step.RetryIndex, step.GroupRetryIdx)
	if s.canceled {
		return nil
	}
	if ctx.Err() != nil {
		return errors.Errorf("step %s cancelled before inner signal dispatch", step.Name)
	}
	// why: Dispatch on a disconnected context: the enqueue commits the inner signal
	// to the DB before returning, so a cancel landing mid-dispatch cannot stop
	// the write — it can only hide the committed result, orphaning a signal
	// whose ID nobody ever learns (failed TestCancelWorkflowPropagatesDown).
	dispatchCtx, dispatchCancel := workflow.NewDisconnectedContext(ctx)
	defer dispatchCancel()
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(dispatchCtx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
		QueueName:       s.TargetQueueName,
		QueueID:         s.TargetQueueID,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
		Callback:        cb,
		DedupeKey:       &dedupeKey,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
	}

	s.innerQueueSignalID = enqueueResp.QueueSignalID

	if s.canceled || ctx.Err() != nil {
		cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
		defer cancelCtxCancel()
		if _, err := client.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID); err != nil {
			if l, logErr := log.WorkflowLogger(ctx); logErr == nil {
				l.Warn("failed to cancel inner signal after mid-dispatch cancel",
					zap.String("step_id", step.ID),
					zap.String("inner_queue_signal_id", enqueueResp.QueueSignalID),
					zap.Error(err))
			}
		}
	}

	logger.Info("waiting for queue signal to complete",
		"step_name", step.Name,
		"queue_signal_id", enqueueResp.QueueSignalID,
	)

	stepTimeout := step.Timeout
	if stepTimeout == 0 {
		stepTimeout = callback.FallbackAwaitTimeout
	}
	_, err = callback.AwaitWithTimeout(ctx, cb, stepTimeout)
	if err != nil {
		return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
	}

	logger.Info("queue signal completed successfully", "step_name", step.Name)
	return nil
}

func isSkippedStatus(status app.Status) bool {
	return status == app.StatusAutoSkipped || status == app.StatusUserSkipped
}
