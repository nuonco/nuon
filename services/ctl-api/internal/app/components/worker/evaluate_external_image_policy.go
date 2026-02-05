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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (w *Workflows) evaluateExternalImagePolicy(ctx workflow.Context, buildID, buildJobID, runnerID, componentName string) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "evaluating image policies")

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return fmt.Errorf("unable to get workflow logger: %w", err)
	}

	l.Info("starting policy evaluation", zap.String("build_id", buildID))

	// Check if any container image policies exist BEFORE fetching metadata
	// This avoids expensive runner job + OCI registry calls when no policies are configured
	policyCheckResult, err := activities.AwaitCheckContainerImagePoliciesExist(ctx, &activities.CheckContainerImagePoliciesExistRequest{
		BuildID: buildID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to check for policies", err))
		return fmt.Errorf("unable to check for container image policies: %w", err)
	}

	if !policyCheckResult.HasPolicies {
		l.Info("no container image policies configured, skipping policy evaluation")
		return nil
	}

	l.Info("container image policies found, proceeding with metadata fetch")

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to get log stream ID", err))
		return fmt.Errorf("unable to get log stream ID: %w", err)
	}

	// Create a fetch-image-metadata job on the runner
	metadataJob, err := activities.AwaitCreateFetchImageMetadataJob(ctx, &activities.CreateFetchImageMetadataJobRequest{
		BuildID:     buildID,
		RunnerID:    runnerID,
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"component_build_id": buildID,
			"build_job_id":       buildJobID,
			"component_name":     componentName,
		},
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to create metadata job", err))
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, "unable to create metadata job")
		return fmt.Errorf("unable to create metadata job: %w", err)
	}

	// Save the job plan
	if err := activities.AwaitSaveFetchImageMetadataPlan(ctx, &activities.SaveFetchImageMetadataPlanRequest{
		JobID:   metadataJob.ID,
		BuildID: buildID,
	}); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to save metadata job plan", err))
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, "unable to save metadata job plan")
		return fmt.Errorf("unable to save metadata job plan: %w", err)
	}

	// Execute the job (queue and poll for completion)
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "fetching image metadata")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   runnerID,
		JobID:      metadataJob.ID,
		WorkflowID: fmt.Sprintf("%s-fetch-image-metadata", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to fetch image metadata", err))
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, "unable to fetch image metadata")
		return fmt.Errorf("unable to fetch image metadata: %w", err)
	}

	// Get the metadata from the job result
	metadataResult, err := activities.AwaitGetImageMetadataFromJobResult(ctx, &activities.GetImageMetadataFromJobResultRequest{
		JobID: metadataJob.ID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to get image metadata", err))
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, "unable to get image metadata")
		return fmt.Errorf("unable to get image metadata: %w", err)
	}

	prepResult, err := activities.AwaitPrepExternalImagePolicy(ctx, &activities.PrepExternalImagePolicyRequest{
		BuildID:       buildID,
		ImageMetadata: metadataResult.Metadata,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to prepare policy evaluation", err))
		w.updateJobStatusForPolicyFailure(ctx, buildJobID, "unable to prepare policy evaluation")
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
			PolicyID:      policy.PolicyID,
			Contents:      policy.Contents,
			InputJSON:     policy.InputJSON,
			InputIndex:    0,
			InputIdentity: policy.InputIdentity,
		})
		futures = append(futures, fut)
	}

	// Collect all violations from parallel evaluations
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

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		l.Warn("unable to get org id", zap.Error(err))
	} else {
		componentID := prepResult.ComponentID
		policyInputCounts := make(map[string]int, len(prepResult.PolicyIDs))
		for _, policyID := range prepResult.PolicyIDs {
			policyInputCounts[policyID] = prepResult.InputCount
		}
		if _, err := sharedactivities.AwaitPersistPolicyReport(ctx, &sharedactivities.PersistPolicyReportRequest{
			OrgID:             orgID,
			AppID:             prepResult.AppID,
			ComponentID:       &componentID,
			InstallSandboxID:  nil,
			OwnerID:           buildID,
			OwnerType:         string(app.PolicyReportOwnerTypeComponentBuild),
			RunnerJobID:       &buildJobID,
			Violations:        allViolations,
			PolicyIDs:         prepResult.PolicyIDs,
			PolicyInputCounts: policyInputCounts,
			OrgName:           prepResult.OrgName,
			AppName:           prepResult.AppName,
			ComponentName:     prepResult.ComponentName,
		}); err != nil {
			l.Warn("failed to persist policy report", zap.Error(err))
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
