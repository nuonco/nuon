package worker

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (w *Workflows) evaluateExternalImagePolicy(ctx workflow.Context, buildID string) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "evaluating image policies")

	metadataResult, err := activities.AwaitFetchImageMetadata(ctx, &activities.FetchImageMetadataRequest{
		BuildID: buildID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to fetch image metadata", err))
		return fmt.Errorf("unable to fetch image metadata: %w", err)
	}

	prepResult, err := activities.AwaitPrepExternalImagePolicy(ctx, &activities.PrepExternalImagePolicyRequest{
		BuildID:       buildID,
		ImageMetadata: metadataResult.Metadata,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage("unable to prepare policy evaluation", err))
		return fmt.Errorf("unable to prepare policy evaluation: %w", err)
	}

	if !prepResult.HasPolicies {
		return nil
	}

	var allViolations []sharedactivities.PolicyViolation
	for _, policy := range prepResult.Policies {
		result, err := sharedactivities.AwaitEvaluateSinglePolicy(ctx, &sharedactivities.EvaluateSinglePolicyRequest{
			PolicyID:  policy.PolicyID,
			Contents:  policy.Contents,
			InputJSON: policy.InputJSON,
		})
		if err != nil {
			w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, truncateErrorMessage(fmt.Sprintf("policy evaluation failed for policy %s", policy.PolicyID), err))
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
		description := formatPolicyViolations("policy violations", denyViolations)
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPolicyFailed, description)
		return fmt.Errorf("image policy check failed: %s", description)
	}

	if len(warnViolations) > 0 {
		description := formatPolicyViolations("policy warnings", warnViolations)
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, description)
	}

	return nil
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
