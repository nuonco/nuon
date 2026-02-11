package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/types/components/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type helmBuildPolicyPayload struct {
	PolicyInput []plan.AdmissionReviewInput `json:"policy_input"`
}

func (w *Workflows) evaluateHelmBuildPolicy(ctx workflow.Context, buildID, buildJobID, componentName string) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "evaluating helm policies")

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return fmt.Errorf("unable to get workflow logger: %w", err)
	}

	l.Info("starting helm policy evaluation", zap.String("build_id", buildID))

	build, err := activities.AwaitGetComponentBuildForPolicyEval(ctx, activities.GetComponentBuildForPolicyEvalRequest{
		ID: buildID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to get build", err))
		return fmt.Errorf("unable to get build: %w", err)
	}

	appConfigID := build.ComponentConfigConnection.AppConfigID
	if appConfigID == "" {
		l.Info("no app config id found, skipping helm policy evaluation")
		return nil
	}

	policiesConfig := build.ComponentConfigConnection.AppConfig.PoliciesConfig
	if policiesConfig.ID == "" {
		l.Info("no policies config found, skipping helm policy evaluation")
		return nil
	}

	applicablePolicies := filterHelmPolicies(policiesConfig.Policies, componentName)
	if len(applicablePolicies) == 0 {
		l.Info("no applicable helm policies found")
		return nil
	}

	jobExec, err := activities.AwaitGetRunnerJobExecutionByJobID(ctx, &activities.GetRunnerJobExecutionByJobIDRequest{
		JobID: buildJobID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to get job execution", err))
		return fmt.Errorf("unable to get runner job execution: %w", err)
	}

	jobExecResult, err := activities.AwaitGetRunnerJobExecutionResult(ctx, &activities.GetRunnerJobExecutionResultRequest{
		RunnerJobExecutionID: jobExec.ID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to get job execution result", err))
		return fmt.Errorf("unable to get runner job execution result: %w", err)
	}

	payload, err := parseHelmBuildPolicyPayload(jobExecResult)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to parse helm policy input", err))
		return fmt.Errorf("unable to parse helm policy input: %w", err)
	}

	if len(payload.PolicyInput) == 0 {
		l.Info("no helm policy inputs found")
		return nil
	}

	policyItems := buildHelmPolicyEvaluationItems(applicablePolicies, payload.PolicyInput)
	if len(policyItems) == 0 {
		l.Info("no helm policy evaluations to run")
		return nil
	}

	policyIDs := make([]string, 0, len(applicablePolicies))
	for _, policy := range applicablePolicies {
		policyIDs = append(policyIDs, policy.ID)
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    1*time.Minute + 30*time.Second,
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	policyCtx := workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for _, policy := range policyItems {
		fut := workflow.ExecuteActivity(policyCtx, (&sharedactivities.Activities{}).EvaluateSinglePolicy, &sharedactivities.EvaluateSinglePolicyRequest{
			PolicyID:      policy.PolicyID,
			Contents:      policy.Contents,
			InputJSON:     policy.InputJSON,
			InputIndex:    policy.InputIndex,
			InputIdentity: policy.InputIdentity,
		})
		futures = append(futures, fut)
	}

	var allViolations []sharedactivities.PolicyViolation
	for _, fut := range futures {
		var result sharedactivities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("policy evaluation failed", err))
			w.updateJobStatusForPolicyFailure(ctx, buildJobID, "policy evaluation failed")
			return fmt.Errorf("policy evaluation failed: %w", err)
		}
		allViolations = append(allViolations, result.Violations...)
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		l.Warn("unable to get org id", zap.Error(err))
	} else {
		policyInputCounts := make(map[string]int, len(policyIDs))
		for _, policyID := range policyIDs {
			policyInputCounts[policyID] = len(payload.PolicyInput)
		}
		componentID := build.ComponentConfigConnection.ComponentID
		if _, err := sharedactivities.AwaitPersistPolicyReport(ctx, &sharedactivities.PersistPolicyReportRequest{
			OrgID:             orgID,
			AppID:             build.ComponentConfigConnection.Component.AppID,
			ComponentID:       &componentID,
			OwnerID:           buildID,
			OwnerType:         string(app.PolicyReportOwnerTypeComponentBuild),
			RunnerJobID:       &buildJobID,
			Violations:        allViolations,
			PolicyIDs:         policyIDs,
			PolicyInputCounts: policyInputCounts,
			OrgName:           build.ComponentConfigConnection.Component.App.Org.Name,
			AppName:           build.ComponentConfigConnection.Component.App.Name,
			ComponentName:     build.ComponentConfigConnection.Component.Name,
		}); err != nil {
			l.Warn("failed to persist policy report", zap.Error(err))
		}
	}

	var denyViolations []sharedactivities.PolicyViolation
	var warnViolations []sharedactivities.PolicyViolation
	for _, v := range allViolations {
		switch v.Severity {
		case "deny":
			denyViolations = append(denyViolations, v)
		case "warn":
			warnViolations = append(warnViolations, v)
		}
	}

	if len(denyViolations) > 0 {
		for _, v := range denyViolations {
			l.Warn("policy violation (deny)", zap.String("message", v.Message))
		}
		description := formatPolicyViolations("policy violations", denyViolations)
		l.Error("policy evaluation failed", zap.Int("deny_count", len(denyViolations)))
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPolicyFailed, description)
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, description)
		return fmt.Errorf("helm policy check failed: %s", description)
	}

	if len(warnViolations) > 0 {
		for _, v := range warnViolations {
			l.Warn("policy violation (warn)", zap.String("message", v.Message))
		}
		description := formatPolicyViolations("policy warnings", warnViolations)
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, description)
	}

	l.Info("helm policy evaluation completed", zap.Int("warn_count", len(warnViolations)))
	return nil
}

func filterHelmPolicies(policies []app.AppPolicyConfig, componentName string) []app.AppPolicyConfig {
	applicable := make([]app.AppPolicyConfig, 0)
	for _, policy := range policies {
		if policy.Engine != config.AppPolicyEngineOPA {
			continue
		}
		if policy.Type != config.AppPolicyTypeHelmChart {
			continue
		}
		if len(policy.Components) == 0 {
			continue
		}
		for _, comp := range policy.Components {
			if comp == "*" || comp == componentName {
				applicable = append(applicable, policy)
				break
			}
		}
	}
	return applicable
}

func parseHelmBuildPolicyPayload(result *app.RunnerJobExecutionResult) (*helmBuildPolicyPayload, error) {
	if result == nil {
		return &helmBuildPolicyPayload{}, nil
	}

	payloadBytes, err := result.GetContentsDecompressedBytes()
	if err != nil {
		return nil, errors.Wrap(err, "unable to decompress helm policy payload")
	}
	if len(payloadBytes) == 0 {
		return &helmBuildPolicyPayload{}, nil
	}

	var payload helmBuildPolicyPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal helm policy payload")
	}

	return &payload, nil
}

func buildHelmPolicyEvaluationItems(policies []app.AppPolicyConfig, inputs []plan.AdmissionReviewInput) []sharedactivities.PolicyToEvaluate {
	result := make([]sharedactivities.PolicyToEvaluate, 0, len(policies)*len(inputs))
	for _, policy := range policies {
		for idx, input := range inputs {
			inputJSON, err := json.Marshal(input)
			if err != nil {
				continue
			}
			result = append(result, sharedactivities.PolicyToEvaluate{
				PolicyID:      policy.ID,
				Contents:      policy.Contents,
				InputJSON:     inputJSON,
				InputIndex:    idx,
				InputIdentity: buildHelmPolicyIdentity(input),
			})
		}
	}
	return result
}

func buildHelmPolicyIdentity(input plan.AdmissionReviewInput) string {
	kind := input.Review.Kind.Kind
	if kind == "" {
		kind = "Unknown"
	}

	metadata, ok := input.Review.Object["metadata"].(map[string]interface{})
	if !ok {
		return kind
	}

	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	if namespace == "" {
		namespace = "default"
	}

	if name == "" {
		return fmt.Sprintf("%s/%s", kind, namespace)
	}

	return fmt.Sprintf("%s/%s/%s", kind, namespace, name)
}
