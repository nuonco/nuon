package policy

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	policyopa "github.com/nuonco/nuon/pkg/policies/opa"
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
	SkipFindingsKey    = "skip_findings"

	policyEvaluationMetric = "policy.evaluation"
)

// Check implements directive.ApprovalCreateCheck for policy evaluation.
type Check struct {
	sig      signal.Signal
	tmw      tmetrics.Writer
	checkCtx *directive.CheckContext
}

func New(sig signal.Signal, tmw tmetrics.Writer, checkCtx *directive.CheckContext) directive.ApprovalCreateCheck {
	return &Check{sig: sig, tmw: tmw, checkCtx: checkCtx}
}

func (c *Check) Name() string { return "policy" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	if flw.PlanOnly {
		return false
	}
	if c.checkCtx.NoopPlan {
		resolve := flw.Metadata[directive.WorkflowMetadataResolveDrift]
		if resolve == nil || *resolve != "true" {
			return false
		}
	}
	if c.requiresEvaluation() {
		return true
	}
	// Signals whose step has no policy target still route their auto-approval
	// decision through this check.
	_, autoApprove := c.sig.(signal.SignalWithAutoApproveOnPoliciesPassing)
	return autoApprove
}

func (c *Check) requiresEvaluation() bool {
	pe, ok := c.sig.(signal.SignalWithPolicyEvaluation)
	return ok && pe.RequiresPolicyEvaluation()
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)

	if !c.requiresEvaluation() {
		result, autoApproved := c.autoApprove(ctx, l, step)
		if autoApproved {
			return result, nil
		}
		return directive.Pass(), nil
	}

	l.Debug("starting policy check",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	violations, skipFindings, checkResults, policyContext, outcome, policyErr := checkPolicies(ctx, step.StepTargetID, step.StepTargetType)
	if policyErr != nil {
		l.Error("failed to check policies",
			zap.String("step_id", step.ID),
			zap.Error(policyErr))
		c.recordFailure(ctx, step, flw, policyErr)
		if policyContext == nil {
			return directive.Pass(), policyErr
		}
	}

	if policyContext == nil {
		l.Debug("no policies apply to workflow step",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		return directive.Pass(), nil
	}

	policyInputCounts := make(map[string]int, len(policyContext.PolicyIDs))
	for _, policyID := range policyContext.PolicyIDs {
		policyInputCounts[policyID] = policyContext.InputCount
	}
	var validationID *string
	if step.PolicyValidation != nil {
		validationID = &step.PolicyValidation.ID
	}
	var runnerJobID *string
	if policyContext.RunnerJobID != "" {
		runnerJobID = &policyContext.RunnerJobID
	}
	reportResult, reportErr := activities.AwaitPersistPolicyReport(ctx, &activities.PersistPolicyReportRequest{
		OrgID:                          policyContext.OrgID,
		AppID:                          policyContext.AppID,
		InstallID:                      policyContext.InstallID,
		InstallSandboxID:               policyContext.InstallSandboxID,
		ComponentID:                    policyContext.ComponentID,
		ComponentBuildID:               policyContext.ComponentBuildID,
		WorkflowStepPolicyValidationID: validationID,
		RunnerJobID:                    runnerJobID,
		RunnerJobExecutionID:           policyContext.RunnerJobExecutionID,
		AppConfigID:                    policyContext.AppConfigID,
		OwnerID:                        step.StepTargetID,
		OwnerType:                      step.StepTargetType,
		Violations:                     violations,
		SkipFindings:                   skipFindings,
		CheckResults:                   checkResults,
		PolicyIDs:                      policyContext.PolicyIDs,
		PolicyInputCounts:              policyInputCounts,
		OrgName:                        policyContext.OrgName,
		AppName:                        policyContext.AppName,
		InstallName:                    policyContext.InstallName,
		ComponentName:                  policyContext.ComponentName,
	})
	if reportErr != nil {
		policyErr := &policyEvaluationError{
			stage: string(policyerrors.EvaluationFailureStagePersistence),
			err:   errors.Wrap(reportErr, "unable to persist policy report"),
		}
		l.Error("failed to persist policy report", zap.Error(reportErr))
		c.recordFailure(ctx, step, flw, policyErr)
		return directive.Pass(), policyErr
	} else if policyErr == nil {
		c.recordSuccess(ctx, step, flw, outcome)
	}

	var passedPolicyIDs []string
	if reportResult != nil {
		passedPolicyIDs = reportResult.PassedPolicyIDs
	}
	if err := processPolicyViolations(ctx, l, step, flw, violations, passedPolicyIDs, skipFindings); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to process check for policy violation")
	}
	if policyErr != nil {
		return directive.Pass(), policyErr
	}

	l.Debug("policy check completed successfully",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if outcome == policyEvaluationOutcomeSkipApply {
		summary := "Apply automatically skipped by policy; existing deployment retained"
		if step.StepTargetType == "install_sandbox_runs" {
			summary = "Sandbox apply automatically skipped by policy; existing sandbox state retained"
		}
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID: step.ID, Status: app.StatusAutoSkipped, StatusDescription: summary,
			Metadata: map[string]any{"policy_action": "skip_apply", "policy_report_id": reportResult.ReportID},
		}); err != nil {
			return directive.Pass(), errors.Wrap(err, "unable to mark deployment as skipped by policy")
		}
		return directive.CheckResult{
			Directive:     directive.StepSkipGroup,
			Status:        app.StatusAutoSkipped,
			SiblingStatus: app.StatusAutoSkipped,
			Reason: directive.CheckReason{
				Check:   "policy-skip-apply",
				Summary: summary,
				Labels:  map[string]string{"policy_report_id": reportResult.ReportID, "policy_action": "skip_apply"},
			},
		}, nil
	}

	// If the signal opts into auto-approve on policies passing and there are
	// no deny violations, short-circuit the pipeline with a continue directive.
	if result, autoApproved := c.autoApprove(ctx, l, step); autoApproved {
		return result, nil
	}

	return directive.Pass(), nil
}

// autoApprove builds the approve-and-continue result for signals that opted into
// auto-approval on passing policies. Callers only reach it once evaluation has
// produced no deny violations, or when the signal has nothing to evaluate.
func (c *Check) autoApprove(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) (directive.CheckResult, bool) {
	aa, ok := c.sig.(signal.SignalWithAutoApproveOnPoliciesPassing)
	if !ok || !aa.AutoApproveOnPoliciesPassing(ctx) {
		return directive.Pass(), false
	}

	if l != nil {
		l.Debug("auto-approving after policies passed",
			zap.String("step_id", step.ID))
	}

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
	}, true
}

type policyEvaluationOutcome string

const (
	policyEvaluationOutcomePass      policyEvaluationOutcome = "pass"
	policyEvaluationOutcomeWarn      policyEvaluationOutcome = "warn"
	policyEvaluationOutcomeDeny      policyEvaluationOutcome = "deny"
	policyEvaluationOutcomeSkipApply policyEvaluationOutcome = "skip_apply"
)

type policyEvaluationError struct {
	stage string
	err   error
}

func (e *policyEvaluationError) Error() string { return e.err.Error() }
func (e *policyEvaluationError) Unwrap() error { return e.err }

func checkPolicies(ctx workflow.Context, stepTargetID, stepTargetType string) ([]activities.PolicyViolation, []activities.PolicyViolation, map[string][]policyopa.Result, *policyhelpers.PolicyEvaluationContext, policyEvaluationOutcome, error) {
	prepResult, err := activities.AwaitPrepPolicyEvaluation(ctx, &activities.PrepPolicyEvaluationRequest{
		StepTargetID:   stepTargetID,
		StepTargetType: stepTargetType,
	})
	if err != nil {
		return nil, nil, nil, nil, "", &policyEvaluationError{
			stage: "preparation",
			err:   errors.Wrap(err, "unable to prepare policy evaluation"),
		}
	}

	if !prepResult.HasPolicies {
		return nil, nil, nil, nil, "", nil
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
			PolicyID:       p.PolicyID,
			PolicyName:     p.PolicyName,
			Contents:       p.Contents,
			InputJSON:      p.InputJSON,
			InputIndex:     p.InputIndex,
			InputIdentity:  p.InputIdentity,
			AllowSkipApply: prepResult.AllowSkipApply,
		})
		futures = append(futures, fut)
	}

	var allViolations, skipFindings []activities.PolicyViolation
	var checkResults map[string][]policyopa.Result
	var evaluationErr error
	outcome := policyEvaluationOutcomePass
	for i, fut := range futures {
		var result activities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			evaluationErr = &policyEvaluationError{
				stage: "evaluation",
				err:   errors.Wrap(err, "policy evaluation failed"),
			}
			if !prepResult.AllowSkipApply {
				return nil, nil, nil, nil, "", evaluationErr
			}
			p := prepResult.Policies[i]
			allViolations = append(allViolations, activities.PolicyViolation{
				PolicyID: p.PolicyID, PolicyName: p.PolicyName,
				InputIndex: p.InputIndex, InputIdentity: p.InputIdentity,
				Severity: "deny", Message: "Policy evaluation failed: " + err.Error(),
			})
			outcome = policyEvaluationOutcomeDeny
			continue
		}
		skipFindings = append(skipFindings, result.SkipFindings...)
		if len(result.Checks) > 0 {
			if checkResults == nil {
				checkResults = make(map[string][]policyopa.Result)
			}
			policyID := prepResult.Policies[i].PolicyID
			checkResults[policyID] = append(checkResults[policyID], result.Checks...)
		}
		for _, check := range result.Checks {
			if check.Decision == "skip_apply" && outcome != policyEvaluationOutcomeDeny {
				outcome = policyEvaluationOutcomeSkipApply
			}
		}
		for _, violation := range result.Violations {
			if violation.Severity == "deny" {
				outcome = policyEvaluationOutcomeDeny
			} else if outcome == policyEvaluationOutcomePass {
				outcome = policyEvaluationOutcomeWarn
			}
		}
		allViolations = append(allViolations, result.Violations...)
	}

	return allViolations, skipFindings, checkResults, &policyhelpers.PolicyEvaluationContext{
		OrgID:                prepResult.OrgID,
		AppID:                prepResult.AppID,
		InstallID:            prepResult.InstallID,
		InstallSandboxID:     prepResult.InstallSandboxID,
		ComponentID:          prepResult.ComponentID,
		ComponentBuildID:     prepResult.ComponentBuildID,
		PolicyIDs:            prepResult.PolicyIDs,
		InputCount:           prepResult.InputCount,
		AppConfigID:          prepResult.AppConfigID,
		RunnerJobID:          prepResult.RunnerJobID,
		RunnerJobExecutionID: prepResult.RunnerJobExecutionID,
		OrgName:              prepResult.OrgName,
		AppName:              prepResult.AppName,
		InstallName:          prepResult.InstallName,
		ComponentName:        prepResult.ComponentName,
	}, outcome, evaluationErr
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

func processPolicyViolations(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, violations []activities.PolicyViolation, passedPolicyIDs []string, skipFindings []activities.PolicyViolation) error {
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
					SkipFindingsKey:    skipFindings,
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
				SkipFindingsKey:    skipFindings,
			},
		},
	}); updateErr != nil {
		if len(skipFindings) > 0 {
			return errors.Wrap(updateErr, "unable to record policy skip findings")
		}
		l.Warn("failed to update step with policy metadata", zap.Error(updateErr))
	}

	return nil
}
