package executeworkflowstep

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/autoapproval"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/noop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/planonly"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/policy"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/checks/staleplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// processPlan runs all plan-related checks for an approval step.
//
// The pipeline has three phases:
//  1. Pre-approval checks (ApprovalCreateCheck) — can short-circuit before user sees the approval
//  2. Await user approval response
//  3. Post-approval checks (ApprovalResponseCheck) — can override the response (e.g. stale plan auto-retry)
func (s *Signal) processPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	l, _ := log.WorkflowLogger(ctx)

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCheckPlan,
			StatusHumanDescription: "checking plan for changes",
			Metadata: map[string]any{
				"status": "checking plan for changes",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step status as checking plan")
	}

	sig := stepSignal(step)
	checkCtx := &directive.CheckContext{}

	// Phase 1: Pre-approval checks
	createChecks := s.approvalCreateChecks(ctx, sig, checkCtx)
	for _, check := range createChecks {
		if s.canceled {
			return nil
		}
		if !check.ShouldRun(step, flw) {
			continue
		}

		result, err := check.Run(ctx, step, flw)
		if err != nil {
			if statusErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"reason": fmt.Sprintf("Step failed during %s.", check.Name()),
					},
					StatusHumanDescription: "Step failed",
				},
			}); statusErr != nil {
				return errors.Wrap(statusErr, "unable to mark step as error")
			}
			return err
		}

		if result.Directive != directive.StepUnknown {
			l.Debug("pre-approval check short-circuited",
				zap.String("check", check.Name()),
				zap.String("directive", string(result.Directive)),
				zap.String("summary", result.Reason.Summary))
			return s.applyCheckResult(ctx, step, flw, result)
		}
	}

	// Phase 2: Await user approval response
	resp, err := s.awaitApprovalResponse(ctx, step, flw)
	if err != nil {
		return err
	}
	if s.retried || s.canceled || s.skipped {
		return nil
	}

	// Phase 3: Post-approval checks (can override the response)
	responseChecks := s.approvalResponseChecks(ctx)
	for _, check := range responseChecks {
		if !check.ShouldRun(step, flw, resp) {
			continue
		}

		result, err := check.Run(ctx, step, flw, resp)
		if err != nil {
			return err
		}

		if result.Directive != directive.StepUnknown {
			l.Debug("post-approval check overrode response",
				zap.String("check", check.Name()),
				zap.String("directive", string(result.Directive)),
				zap.String("summary", result.Reason.Summary))
			return s.applyCheckResult(ctx, step, flw, result)
		}
	}

	// Phase 4: Normal response handling
	return s.dispatchApprovalResponse(ctx, step, flw, resp)
}

// approvalCreateChecks returns the ordered list of pre-approval checks.
func (s *Signal) approvalCreateChecks(ctx workflow.Context, sig qsignal.Signal, checkCtx *directive.CheckContext) []directive.ApprovalCreateCheck {
	// Load org feature flags needed by checks. Best-effort: default false on error.
	orgAutoSkipNoop, _ := activities.AwaitCheckOrgFeatureByFeature(ctx, string(app.OrgFeatureAutoSkipNoop))
	return []directive.ApprovalCreateCheck{
		noop.New(sig, checkCtx, orgAutoSkipNoop, setResultDirective),
		policy.New(sig),
		autoapproval.New(stepSignal, setResultDirective),
		planonly.New(s.OwnerID, checkCtx),
	}
}

// approvalResponseChecks returns the ordered list of post-approval checks.
func (s *Signal) approvalResponseChecks(ctx workflow.Context) []directive.ApprovalResponseCheck {
	// Load configurable stale plan threshold. Best-effort: empty string = use default.
	thresholdStr, _ := activities.AwaitGetStalePlanThreshold(ctx, activities.GetStalePlanThresholdRequest{})
	var threshold time.Duration
	if thresholdStr != "" {
		threshold, _ = time.ParseDuration(thresholdStr)
	}
	return []directive.ApprovalResponseCheck{
		staleplan.New(threshold, setResultDirective),
	}
}

// applyCheckResult writes the directive and reason metadata from a check result.
func (s *Signal) applyCheckResult(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, result directive.CheckResult) error {
	meta := result.Reason.Metadata()
	meta[string(directive.MetadataKey)] = string(result.Directive)

	if err := setResultDirective(ctx, step.ID, result.Directive); err != nil {
		return errors.Wrap(err, "unable to set result directive from check")
	}

	stepStatus := result.Status
	if stepStatus == "" {
		stepStatus = app.StatusError
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 stepStatus,
			StatusHumanDescription: result.Reason.Summary,
			Metadata:               meta,
		},
	})

	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: result.Reason.Summary,
			Metadata: map[string]any{
				"step_id": step.ID,
				"check":   result.Reason.Check,
			},
		},
	})

	return nil
}
