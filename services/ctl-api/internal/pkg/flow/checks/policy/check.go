package policy

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	policyhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/policyerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const (
	DenyViolationsKey  = "deny_violations"
	WarnViolationsKey  = "warn_violations"
	PassedPolicyIDsKey = "passed_policy_ids"

	policyEvaluationMetric = "policy.evaluation"
)

// Check implements directive.ApprovalCreateCheck for policy evaluation.
type Check struct {
	sig signal.Signal
	tmw tmetrics.Writer
}

func New(sig signal.Signal, tmw tmetrics.Writer) directive.ApprovalCreateCheck {
	return &Check{sig: sig, tmw: tmw}
}

func (c *Check) Name() string { return "policy" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	pe, ok := c.sig.(signal.SignalWithPolicyEvaluation)
	return ok && pe.RequiresPolicyEvaluation()
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)

	l.Debug("starting policy check",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	violations, policyContext, outcome, policyErr := checkPolicies(ctx, step.StepTargetID, step.StepTargetType)
	if policyErr != nil {
		l.Error("failed to check policies; continuing with warning",
			zap.String("step_id", step.ID),
			zap.Error(policyErr))
		c.recordFailure(ctx, step, flw, policyErr)
		recordEvaluationCompositeError(ctx, l, step, policyErr)
		return directive.Pass(), nil
	}

	if policyContext != nil {
		c.recordSuccess(ctx, step, flw, outcome)

		policyInputCounts := make(map[string]int, len(policyContext.PolicyIDs))
		for _, policyID := range policyContext.PolicyIDs {
			policyInputCounts[policyID] = policyContext.InputCount
		}
		var validationID *string
		if step.PolicyValidation != nil {
			validationID = &step.PolicyValidation.ID
		}
		reportResult, err := activities.AwaitPersistPolicyReport(ctx, &activities.PersistPolicyReportRequest{
			OrgID:                          policyContext.OrgID,
			AppID:                          policyContext.AppID,
			InstallID:                      policyContext.InstallID,
			InstallSandboxID:               policyContext.InstallSandboxID,
			ComponentID:                    policyContext.ComponentID,
			ComponentBuildID:               policyContext.ComponentBuildID,
			WorkflowStepPolicyValidationID: validationID,
			OwnerID:                        step.StepTargetID,
			OwnerType:                      step.StepTargetType,
			Violations:                     violations,
			PolicyIDs:                      policyContext.PolicyIDs,
			PolicyInputCounts:              policyInputCounts,
			OrgName:                        policyContext.OrgName,
			AppName:                        policyContext.AppName,
			InstallName:                    policyContext.InstallName,
			ComponentName:                  policyContext.ComponentName,
		})
		if err != nil {
			l.Warn("failed to persist policy report", zap.Error(err))
		}

		var passedPolicyIDs []string
		if reportResult != nil {
			passedPolicyIDs = reportResult.PassedPolicyIDs
		}
		if err := processPolicyViolations(ctx, l, step, flw, violations, passedPolicyIDs); err != nil {
			return directive.Pass(), errors.Wrap(err, "unable to process check for policy violation")
		}
	}

	l.Debug("policy check completed successfully",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	// If the signal opts into auto-approve on policies passing and there are
	// no deny violations, short-circuit the pipeline with a continue directive.
	if aa, ok := c.sig.(signal.SignalWithAutoApproveOnPoliciesPassing); ok && aa.AutoApproveOnPoliciesPassing(ctx) {
		l.Debug("auto-approving after policies passed",
			zap.String("step_id", step.ID))
		return directive.CheckResult{
			Directive: directive.StepContinue,
			Status:    app.WorkflowStepApprovalStatusApproved,
			Reason: directive.CheckReason{
				Check:   "policy-auto-approve",
				Summary: "Auto-approved: all policies passed",
				Labels: map[string]string{
					"auto_approved":   "true",
					"approval_reason": "policies_passed",
				},
			},
		}, nil
	}

	return directive.Pass(), nil
}

type policyEvaluationOutcome string

const (
	policyEvaluationOutcomePass policyEvaluationOutcome = "pass"
	policyEvaluationOutcomeWarn policyEvaluationOutcome = "warn"
	policyEvaluationOutcomeDeny policyEvaluationOutcome = "deny"
)

type policyEvaluationError struct {
	stage string
	err   error
}

func (e *policyEvaluationError) Error() string { return e.err.Error() }
func (e *policyEvaluationError) Unwrap() error { return e.err }

func checkPolicies(ctx workflow.Context, stepTargetID, stepTargetType string) ([]activities.PolicyViolation, *policyhelpers.PolicyEvaluationContext, policyEvaluationOutcome, error) {
	prepResult, err := activities.AwaitPrepPolicyEvaluation(ctx, &activities.PrepPolicyEvaluationRequest{
		StepTargetID:   stepTargetID,
		StepTargetType: stepTargetType,
	})
	if err != nil {
		return nil, nil, "", &policyEvaluationError{
			stage: "preparation",
			err:   errors.Wrap(err, "unable to prepare policy evaluation"),
		}
	}

	if !prepResult.HasPolicies {
		return nil, nil, "", nil
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    1*time.Minute + 30*time.Second,
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	policyCtx := workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for _, p := range prepResult.Policies {
		fut := workflow.ExecuteActivity(policyCtx, (&activities.Activities{}).EvaluateSinglePolicy, &activities.EvaluateSinglePolicyRequest{
			PolicyID:      p.PolicyID,
			PolicyName:    p.PolicyName,
			Contents:      p.Contents,
			InputJSON:     p.InputJSON,
			InputIndex:    p.InputIndex,
			InputIdentity: p.InputIdentity,
		})
		futures = append(futures, fut)
	}

	var allViolations []activities.PolicyViolation
	outcome := policyEvaluationOutcomePass
	for _, fut := range futures {
		var result activities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			return nil, nil, "", &policyEvaluationError{
				stage: "evaluation",
				err:   errors.Wrap(err, "policy evaluation failed"),
			}
		}
		for _, violation := range result.Violations {
			if violation.Severity == "deny" {
				outcome = policyEvaluationOutcomeDeny
			} else if outcome != policyEvaluationOutcomeDeny {
				outcome = policyEvaluationOutcomeWarn
			}
		}
		allViolations = append(allViolations, result.Violations...)
	}

	return allViolations, &policyhelpers.PolicyEvaluationContext{
		OrgID:            prepResult.OrgID,
		AppID:            prepResult.AppID,
		InstallID:        prepResult.InstallID,
		InstallSandboxID: prepResult.InstallSandboxID,
		ComponentID:      prepResult.ComponentID,
		ComponentBuildID: prepResult.ComponentBuildID,
		PolicyIDs:        prepResult.PolicyIDs,
		InputCount:       prepResult.InputCount,
		OrgName:          prepResult.OrgName,
		AppName:          prepResult.AppName,
		InstallName:      prepResult.InstallName,
		ComponentName:    prepResult.ComponentName,
	}, outcome, nil
}

func (c *Check) recordFailure(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, policyErr error) {
	if c.tmw == nil {
		return
	}

	c.tmw.Incr(ctx, policyEvaluationMetric, metrics.ToTags(map[string]string{
		"org_id":           flw.OrgID,
		"stage":            policyEvaluationErrorStage(policyErr),
		"status":           "error",
		"step_target_type": step.StepTargetType,
		"workflow_type":    string(flw.Type),
	})...)
}

func policyEvaluationErrorStage(policyErr error) string {
	var evaluationErr *policyEvaluationError
	if errors.As(policyErr, &evaluationErr) {
		return evaluationErr.stage
	}
	return "unknown"
}

func recordEvaluationCompositeError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, policyErr error) {
	if err := activities.AwaitRecordPolicyEvaluationCompositeError(ctx, activities.RecordPolicyEvaluationCompositeErrorRequest{
		WorkflowStepID: step.ID,
		StepTargetID:   step.StepTargetID,
		StepTargetType: step.StepTargetType,
		Stage:          policyerrors.EvaluationFailureStage(policyEvaluationErrorStage(policyErr)),
	}); err != nil {
		l.Warn("failed to record policy evaluation composite error",
			zap.String("step_id", step.ID),
			zap.Error(err))
	}
}

func (c *Check) recordSuccess(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, outcome policyEvaluationOutcome) {
	if c.tmw == nil {
		return
	}

	tags := map[string]string{
		"org_id":           flw.OrgID,
		"status":           "success",
		"step_target_type": step.StepTargetType,
		"workflow_type":    string(flw.Type),
	}
	c.tmw.Incr(ctx, policyEvaluationMetric, metrics.ToTags(tags, "outcome", string(outcome))...)
}

func processPolicyViolations(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, violations []activities.PolicyViolation, passedPolicyIDs []string) error {
	var denyViolations, warnViolations []activities.PolicyViolation
	for _, v := range violations {
		if v.Severity == "deny" {
			denyViolations = append(denyViolations, v)
		} else {
			warnViolations = append(warnViolations, v)
		}
	}

	l.Info("policy evaluation complete",
		zap.String("step_id", step.ID),
		zap.Int("deny_count", len(denyViolations)),
		zap.Int("warn_count", len(warnViolations)),
		zap.Int("passed_count", len(passedPolicyIDs)))

	if len(denyViolations) > 0 {
		if updateErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusError,
				Metadata: map[string]any{
					"reason":           "Policy violations found",
					DenyViolationsKey:  denyViolations,
					WarnViolationsKey:  warnViolations,
					PassedPolicyIDsKey: passedPolicyIDs,
				},
				StatusHumanDescription: "Policy check failed",
			},
		}); updateErr != nil {
			return errors.Wrap(updateErr, "unable to mark step as error")
		}
		return fmt.Errorf("policy violations found: %d deny violations", len(denyViolations))
	}

	if updateErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: step.Status.Status,
			Metadata: map[string]any{
				DenyViolationsKey:  denyViolations,
				WarnViolationsKey:  warnViolations,
				PassedPolicyIDsKey: passedPolicyIDs,
			},
		},
	}); updateErr != nil {
		l.Warn("failed to update step with policy metadata", zap.Error(updateErr))
	}

	return nil
}
