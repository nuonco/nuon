package autoapproval

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Check implements directive.ApprovalCreateCheck for auto-approval.
// Skipped in plan-only mode so planonly.Check owns the approval semantics.
type Check struct {
	// StepSignalFn returns the inner signal from a step for interface assertions.
	StepSignalFn func(step *app.WorkflowStep) signal.Signal

	// SetResultDirective writes the directive to the step's ResultDirective column.
	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(stepSignalFn func(*app.WorkflowStep) signal.Signal, setDirective func(workflow.Context, string, directive.Step) error) directive.ApprovalCreateCheck {
	return &Check{StepSignalFn: stepSignalFn, SetResultDirective: setDirective}
}

func (c *Check) Name() string { return "auto-approval" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	return !flw.PlanOnly
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	if flw.ApprovalOption != app.InstallApprovalOptionApproveAll {
		return directive.Pass(), nil
	}

	approvalType := "unknown"
	if at, ok := flw.Status.Metadata["approval_type"]; ok {
		if s, ok := at.(string); ok {
			approvalType = s
		}
	}

	_, err := activities.AwaitCreateApprovalResponse(ctx, activities.CreateStepApprovalResponseRequest{
		StepApprovalID: step.Approval.ID,
		Type:           app.WorkflowStepApprovalResponseTypeApprove,
		Note:           "auto-approved",
	})
	if err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to create auto-approval response")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto-approved " + step.Name,
			Metadata: map[string]any{
				"step_idx":      step.Idx,
				"status":        "auto-approved",
				"auto_approved": true,
				"approval_type": approvalType,
			},
		},
	}); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to update step status for auto-approval")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "auto-approved step " + step.Name,
			Metadata: map[string]any{
				"step_idx":      step.Idx,
				"status":        "auto-approved",
				"approval_type": approvalType,
			},
		},
	}); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to update flow status for auto-approval")
	}

	if sig := c.StepSignalFn(step); sig != nil {
		if oa, ok := sig.(signal.SignalWithOnApprove); ok {
			if err := oa.OnApprove(ctx); err != nil {
				l, _ := log.WorkflowLogger(ctx)
				if l != nil {
					l.Warn("OnApprove hook failed during auto-approval", zap.Error(err))
				}
			}
		}
	}

	if err := c.SetResultDirective(ctx, step.ID, directive.StepContinue); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to set result directive for auto-approval")
	}

	return directive.CheckResult{
		Directive: directive.StepContinue,
		Status:    app.WorkflowStepApprovalStatusApproved,
		Reason: directive.CheckReason{
			Check:   "auto-approval",
			Summary: "Auto-approved (approve-all enabled)",
			Labels: map[string]string{
				"auto_approved":   "true",
				"approval_type":   approvalType,
				"approval_reason": "approve_all",
			},
		},
	}, nil
}
