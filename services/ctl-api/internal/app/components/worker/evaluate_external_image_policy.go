package worker

import (
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (w *Workflows) evaluateExternalImagePolicy(ctx workflow.Context, buildID, jobID string) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "evaluating image policies")

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return fmt.Errorf("unable to get workflow logger: %w", err)
	}

	l.Info("starting policy evaluation", zap.String("build_id", buildID))

	metadataResult, err := activities.AwaitFetchImageMetadata(ctx, &activities.FetchImageMetadataRequest{
		BuildID: buildID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to fetch image metadata", err))
		w.updateJobStatusForPolicyFailure(ctx, jobID, "unable to fetch image metadata")
		return fmt.Errorf("unable to fetch image metadata: %w", err)
	}

	prepResult, err := activities.AwaitPrepExternalImagePolicy(ctx, &activities.PrepExternalImagePolicyRequest{
		BuildID:       buildID,
		ImageMetadata: metadataResult.Metadata,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to prepare policy evaluation", err))
		w.updateJobStatusForPolicyFailure(ctx, jobID, "unable to prepare policy evaluation")
		return fmt.Errorf("unable to prepare policy evaluation: %w", err)
	}

	if !prepResult.HasPolicies {
		l.Info("no policies configured, skipping policy evaluation")
		return nil
	}

	l.Info("evaluating policies", zap.Int("policy_count", len(prepResult.Policies)))

	// Execute all policy evaluations in parallel using futures
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    1*time.Minute + 30*time.Second,
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	policyCtx := workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for _, policy := range prepResult.Policies {
		fut := workflow.ExecuteActivity(policyCtx, (&sharedactivities.Activities{}).EvaluateSinglePolicy, &sharedactivities.EvaluateSinglePolicyRequest{
			PolicyID:  policy.PolicyID,
			Contents:  policy.Contents,
			InputJSON: policy.InputJSON,
		})
		futures = append(futures, fut)
	}

	// Collect all violations from parallel evaluations
	var allViolations []sharedactivities.PolicyViolation
	for _, fut := range futures {
		var result sharedactivities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("policy evaluation failed", err))
			w.updateJobStatusForPolicyFailure(ctx, jobID, "policy evaluation failed")
			return fmt.Errorf("policy evaluation failed: %w", err)
		}
		allViolations = append(allViolations, result.Violations...)
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
		w.updateJobStatusForPolicyFailure(ctx, jobID, description)
		return fmt.Errorf("image policy check failed: %s", description)
	}

	if len(warnViolations) > 0 {
		for _, v := range warnViolations {
			l.Warn("policy violation (warn)", zap.String("message", v.Message))
		}
		description := formatPolicyViolations("policy warnings", warnViolations)
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, description)
	}

	l.Info("policy evaluation completed", zap.Int("warn_count", len(warnViolations)))
	return nil
}

func (w *Workflows) updateJobStatusForPolicyFailure(ctx workflow.Context, jobID, description string) {
	_ = activities.AwaitUpdateJobStatus(ctx, &activities.UpdateJobStatusRequest{
		JobID:             jobID,
		Status:            app.RunnerJobStatusFailed,
		StatusDescription: description,
	})
}

const maxDescriptionLength = 500

func truncateErrorMessage(prefix string, err error) string {
	if err == nil {
		return prefix
	}
	msg := fmt.Sprintf("%s: %v", prefix, err)
	if len(msg) > maxDescriptionLength {
		return msg[:maxDescriptionLength-3] + "..."
	}
	return msg
}

func formatPolicyViolations(prefix string, violations []sharedactivities.PolicyViolation) string {
	prefix = prefix + ": "
	const separator = "; "

	if len(violations) == 0 {
		return prefix + "none"
	}

	var result strings.Builder
	result.WriteString(prefix)

	includedCount := 0
	for i, v := range violations {
		msg := v.Message
		if i > 0 {
			msg = separator + msg
		}

		suffix := ""
		remaining := len(violations) - includedCount - 1
		if remaining > 0 {
			suffix = fmt.Sprintf("... (+%d more)", remaining)
		}

		if result.Len()+len(msg)+len(suffix) > maxDescriptionLength && includedCount > 0 {
			result.WriteString(fmt.Sprintf("... (+%d more)", len(violations)-includedCount))
			break
		}

		result.WriteString(msg)
		includedCount++
	}

	return result.String()
}
