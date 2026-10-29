package autoapproval

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/staleplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/superseded"
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
	if directive.ResolveDrift(ctx, flw) {
		if !directive.ResolveDriftApprovalPolicy(ctx, flw) || flw.ApprovalOption != app.InstallApprovalOptionApproveAll {
			return c.runPriorApprovalMatch(ctx, step, flw)
		}
	}

	if flw.ApprovalOption != app.InstallApprovalOptionApproveAll {
		return directive.Pass(), nil
	}

	approvalType := "unknown"
	if at, ok := flw.Status.Metadata["approval_type"]; ok {
		if s, ok := at.(string); ok {
			approvalType = s
		}
	}

	return c.approve(ctx, step, flw, autoApproval{
		note:                   "auto-approved",
		approvalType:           approvalType,
		summary:                "Auto-approved (approve-all enabled)",
		reason:                 "approve_all",
		requireApproveResponse: directive.ResolveDriftApprovalPolicy(ctx, flw),
	})
}

type autoApproval struct {
	note         string
	approvalType string
	summary      string
	reason       string
	labels       map[string]string

	requireApproveResponse bool
}

func (c *Check) runPriorApprovalMatch(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)
	priorApprovalID := directive.ReuseApprovalID(flw)
	if priorApprovalID == "" || step.Approval == nil {
		return directive.Pass(), nil
	}

	match, err := activities.AwaitMatchPriorApprovalPlan(ctx, &activities.MatchPriorApprovalPlanRequest{
		OrgID:           flw.OrgID,
		InstallID:       flw.OwnerID,
		WorkflowID:      flw.ID,
		StepID:          step.ID,
		ApprovalID:      step.Approval.ID,
		PriorApprovalID: priorApprovalID,
	})
	if err != nil {
		if l != nil {
			l.Warn("unable to match prior approval plan; awaiting approval",
				zap.String("step_id", step.ID),
				zap.String("prior_approval_id", priorApprovalID),
				zap.Error(err))
		}
		return directive.Pass(), nil
	}
	if !match.Matched {
		if l != nil {
			l.Info("plan does not match prior approval; awaiting approval",
				zap.String("step_id", step.ID),
				zap.String("prior_approval_id", priorApprovalID),
				zap.String("reason", match.Reason))
		}
		return directive.Pass(), nil
	}

	thresholdStr, err := activities.AwaitGetStalePlanThreshold(ctx, activities.GetStalePlanThresholdRequest{})
	if err != nil {
		return directive.Pass(), nil
	}
	threshold, _ := time.ParseDuration(thresholdStr)
	response := &app.WorkflowStepApprovalResponse{
		Type:      app.WorkflowStepApprovalResponseTypeApprove,
		CreatedAt: workflow.Now(ctx),
	}
	for _, check := range []directive.ApprovalResponseCheck{
		staleplan.New(threshold, c.SetResultDirective),
		superseded.New(c.StepSignalFn, c.SetResultDirective),
	} {
		if !check.ShouldRun(step, flw, response) {
			continue
		}
		result, err := check.Run(ctx, step, flw, response)
		if err != nil || result.Directive != directive.StepUnknown {
			return result, err
		}
	}

	return c.approve(ctx, step, flw, autoApproval{
		note:                   "auto-approved: plan matches approval " + priorApprovalID,
		approvalType:           "prior-approval",
		summary:                "Auto-approved (plan matches prior approval)",
		reason:                 "prior_approval_match",
		requireApproveResponse: true,
		labels: map[string]string{
			"resolve_drift":     "true",
			"prior_approval_id": priorApprovalID,
			"prior_response_id": match.PriorResponseID,
		},
	})
}

func (c *Check) approve(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, aa autoApproval) (directive.CheckResult, error) {
	approvalType := aa.approvalType

	resp, err := activities.AwaitCreateApprovalResponse(ctx, activities.CreateStepApprovalResponseRequest{
		StepApprovalID: step.Approval.ID,
		Type:           app.WorkflowStepApprovalResponseTypeApprove,
		Note:           aa.note,
	})
	if err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to create auto-approval response")
	}
	if aa.requireApproveResponse && (resp == nil || resp.Type != app.WorkflowStepApprovalResponseTypeApprove) {
		return directive.Pass(), nil
	}

	stepMeta := map[string]any{
		"step_idx":      step.Idx,
		"status":        "auto-approved",
		"auto_approved": true,
		"approval_type": approvalType,
	}
	for k, v := range aa.labels {
		stepMeta[k] = v
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto-approved " + step.Name,
			Metadata:               stepMeta,
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

	labels := map[string]string{
		"auto_approved":   "true",
		"approval_type":   approvalType,
		"approval_reason": aa.reason,
	}
	for k, v := range aa.labels {
		labels[k] = v
	}

	return directive.CheckResult{
		Directive: directive.StepContinue,
		Status:    app.WorkflowStepApprovalStatusApproved,
		Reason: directive.CheckReason{
			Check:   "auto-approval",
			Summary: aa.summary,
			Labels:  labels,
		},
	}, nil
}
