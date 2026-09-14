package statusactivities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type StartFlowStepRequest struct {
	StepID   string `validate:"required"`
	FlowID   string `validate:"required"`
	StepName string
}

type FinishFlowStepRequest struct {
	StepID   string `validate:"required"`
	FlowID   string `validate:"required"`
	StepName string
	StepIdx  int
}

// PkgStatusStartFlowStep performs the three writes that open a step — the flow's
// "executing step" status, the step's started_at, and the step's in-progress
// status — in one activity instead of three. Each was its own round trip, and at
// ~115ms of Temporal dispatch per round trip that dominated the step's own work.
//
// @temporal-gen-v2 activity
// @local
// @local-retry-policy-max-attempts 3
func (a *Activities) PkgStatusStartFlowStep(ctx context.Context, req StartFlowStepRequest) error {
	if err := a.PkgStatusUpdateFlowStatus(ctx, UpdateStatusRequest{
		ID: req.FlowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + req.StepName,
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update flow status")
	}

	if err := a.setFlowStepStartedAt(ctx, req.StepID); err != nil {
		return err
	}

	if err := a.PkgStatusUpdateFlowStepStatus(ctx, UpdateStatusRequest{
		ID:     req.StepID,
		Status: app.CompositeStatus{Status: app.StatusInProgress},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step in progress")
	}

	return nil
}

// PkgStatusFinishFlowStep closes out a successful non-approval step: it marks the
// step successful and writes the flow's "finished executing step" status, in one
// activity instead of two round trips plus the caller's own re-read.
//
// @temporal-gen-v2 activity
// @local
// @local-retry-policy-max-attempts 3
func (a *Activities) PkgStatusFinishFlowStep(ctx context.Context, req FinishFlowStepRequest) error {
	var step app.WorkflowStep
	if res := a.db.WithContext(ctx).Where(app.WorkflowStep{ID: req.StepID}).First(&step); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get step")
	}

	// A signal that skipped its own work marks the step skipped before
	// returning. Overwriting that with success would report work as done that
	// never ran, so leave an already-skipped status alone.
	if !isSkippedStatus(step.Status.Status) {
		if err := a.PkgStatusUpdateFlowStepStatus(ctx, UpdateStatusRequest{
			ID:     req.StepID,
			Status: app.CompositeStatus{Status: app.StatusSuccess},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as success")
		}
	}

	if err := a.PkgStatusUpdateFlowStatus(ctx, UpdateStatusRequest{
		ID: req.FlowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "finished executing step " + req.StepName,
			Metadata: map[string]any{
				"step_idx": req.StepIdx,
				"status":   "ok",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update flow status after step")
	}

	return nil
}

// setFlowStepStartedAt stamps started_at only when it is still null. A local
// activity can run more than once — a workflow task that fails mid-flight
// replays every local activity it had not yet committed — and an unconditional
// write would move the timestamp, shrinking the duration the UI reports.
func (a *Activities) setFlowStepStartedAt(ctx context.Context, stepID string) error {
	res := a.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Where(app.WorkflowStep{ID: stepID}).
		Where("started_at IS NULL").
		Update("started_at", time.Now())
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to set step started at")
	}

	return nil
}

func isSkippedStatus(status app.Status) bool {
	return status == app.StatusAutoSkipped || status == app.StatusUserSkipped
}
