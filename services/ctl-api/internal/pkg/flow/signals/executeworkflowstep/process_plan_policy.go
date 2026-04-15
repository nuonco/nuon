package executeworkflowstep

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	policyhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// runPolicyCheck evaluates policies and processes results.
// Always returns done=false (policy checks never short-circuit the plan flow).
func (s *Signal) runPolicyCheck(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) (bool, error) {
	l.Debug("starting policy check",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	violations, policyContext, policyErr := s.checkPolicies(ctx, step.StepTargetID, step.StepTargetType)
	if policyErr != nil {
		l.Warn("failed to check policies",
			zap.String("step_id", step.ID),
			zap.String("step_target_id", step.StepTargetID),
			zap.String("step_target_type", step.StepTargetType),
			zap.String("workflow_id", flw.ID),
			zap.Error(policyErr))
	}

	if policyContext != nil {
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
		if err := s.processPolicyViolations(ctx, l, step, flw, violations, passedPolicyIDs); err != nil {
			return false, errors.Wrap(err, "unable to process check for policy violation")
		}
	}

	l.Debug("policy check completed successfully",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	return false, nil
}

// checkPolicies evaluates all applicable policies for the step target.
func (s *Signal) checkPolicies(ctx workflow.Context, stepTargetID, stepTargetType string) ([]activities.PolicyViolation, *policyhelpers.PolicyEvaluationContext, error) {
	prepResult, err := activities.AwaitPrepPolicyEvaluation(ctx, &activities.PrepPolicyEvaluationRequest{
		StepTargetID:   stepTargetID,
		StepTargetType: stepTargetType,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to prepare policy evaluation")
	}

	if !prepResult.HasPolicies {
		return nil, nil, nil
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    1*time.Minute + 30*time.Second,
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	policyCtx := workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for _, policy := range prepResult.Policies {
		fut := workflow.ExecuteActivity(policyCtx, (&activities.Activities{}).EvaluateSinglePolicy, &activities.EvaluateSinglePolicyRequest{
			PolicyID:      policy.PolicyID,
			PolicyName:    policy.PolicyName,
			Contents:      policy.Contents,
			InputJSON:     policy.InputJSON,
			InputIndex:    policy.InputIndex,
			InputIdentity: policy.InputIdentity,
		})
		futures = append(futures, fut)
	}

	var allViolations []activities.PolicyViolation
	for _, fut := range futures {
		var result activities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			return nil, nil, errors.Wrap(err, "policy evaluation failed")
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
	}, nil
}

// processPolicyViolations handles policy evaluation results.
func (s *Signal) processPolicyViolations(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, violations []activities.PolicyViolation, passedPolicyIDs []string) error {
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
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID),
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
