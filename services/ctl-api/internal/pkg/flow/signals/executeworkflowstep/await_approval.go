package executeworkflowstep

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (s *Signal) awaitApprovalResponse(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (*app.WorkflowStepApprovalResponse, error) {
	if err := setResultDirective(ctx, step.ID, DirectiveAwaitApproval); err != nil {
		return nil, errors.Wrap(err, "unable to write await-approval directive")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_idx":           step.Idx,
				"status":             "ok",
				string(DirectiveKey): string(DirectiveAwaitApproval),
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to update step to awaiting approval status")
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_id": step.ID,
			},
		},
	})

	if s.ResidentFlow {
		return nil, errApprovalParked
	}

	return s.waitForApprovalResponse(ctx, flw, step)
}

var errApprovalParked = errors.New("approval parked")

func (s *Signal) resumeApproval(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	if step.Status.Status != app.AwaitingApproval || step.Approval == nil || step.Approval.Response == nil {
		l.Debug("step is not parked awaiting approval with a response, exiting",
			zap.String("step_id", step.ID),
			zap.String("step_status", string(step.Status.Status)))
		return nil
	}

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, step.ID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	return s.processApprovalResponse(ctx, step, flw, step.Approval.Response)
}

func (s *Signal) dispatchApprovalResponse(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) error {
	l, _ := log.WorkflowLogger(ctx)

	switch resp.Type {
	case app.WorkflowStepApprovalResponseTypeApprove:
		return s.handleApproveResponse(ctx, l, step, flw)
	case app.WorkflowStepApprovalResponseTypeSkipCurrent:
		return s.handleSkipResponse(ctx, l, step, flw)
	case app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents:
		return s.handleSkipDependentsResponse(ctx, l, step, flw)
	default:
		return s.handleDenyResponse(ctx, l, step, flw)
	}
}

var errApprovalExpired = errors.New("approval expired")

// why: approvalExpiredStopVersion gates the expired-approval stop path: histories
// written before it recorded only the step-status update on expiry, so the
// target-status and stop-directive activities must not replay into them.
const approvalExpiredStopVersion = "approval-expired-stop-v1"

func (s *Signal) waitForApprovalResponse(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) (*app.WorkflowStepApprovalResponse, error) {
	ok, err := workflow.AwaitWithTimeout(ctx, callback.MaxWaitCeiling, func() bool {
		return s.approved || s.retried || s.canceled || s.skipped
	})
	if err != nil {
		return nil, fmt.Errorf("error waiting for approval for step %s: %w", step.ID, err)
	}
	if !ok {
		expired := app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalExpired, map[string]any{
			"err_message": "approval was not accepted",
		})
		expired.StatusHumanDescription = "no approval received"
		statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     step.ID,
			Status: expired,
		})
		if workflow.GetVersion(ctx, approvalExpiredStopVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
			return nil, fmt.Errorf("approval timed out for step %s", step.ID)
		}
		if terr := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID:            step.ID,
			Status:            app.WorkflowStepApprovalStatusApprovalExpired,
			StatusDescription: "no approval received",
		}); terr != nil {
			return nil, errors.Wrap(terr, "unable to update step target status for expired approval")
		}
		// why: Stop the group instead of erroring: an expired approval must not
		// enter the retry machinery and re-park.
		if derr := setResultDirective(ctx, step.ID, DirectiveStop); derr != nil {
			return nil, errors.Wrap(derr, "unable to set stop directive for expired approval")
		}
		return nil, errApprovalExpired
	}

	if s.retried || s.canceled || s.skipped {
		return nil, nil
	}

	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step after approval")
	}
	if step.Approval == nil || step.Approval.Response == nil {
		return nil, errors.New("approval response not found after update")
	}
	return step.Approval.Response, nil
}

func (s *Signal) awaitAndHandleApproval(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	l, _ := log.WorkflowLogger(ctx)
	_ = l

	resp, err := s.awaitApprovalResponse(ctx, step, flw)
	if err != nil {
		return err
	}

	if s.retried || s.canceled {
		return nil
	}

	return s.dispatchApprovalResponse(ctx, step, flw, resp)
}
