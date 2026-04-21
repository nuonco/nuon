package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Directive constants — duplicated here to avoid import cycle with the flow package.
// The canonical definitions are in flow.DirectiveX.
const (
	DirectiveKey           = "directive"
	DirectiveContinue      = "continue"
	DirectiveSkipGroup     = "skip-group"
	DirectiveStop          = "stop"
	DirectiveRetry         = "retry"
	DirectiveRetryGroup    = "retry-group"
	DirectiveAwaitApproval = "await-approval"
)

// setResultDirective writes only the ResultDirective column on the step.
// Use this when the step status should NOT be changed (e.g., the step failed
// and we want to keep it as StatusError while signaling retry-group).
func setResultDirective(ctx workflow.Context, stepID string, directive string) error {
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID:    stepID,
		Directive: directive,
	})
}

// writeDirective writes a directive to both the step's ResultDirective column
// AND marks the step status as StatusSuccess with the directive in metadata.
// Use this for normal completion directives (continue, stop, skip-group, await-approval)
// where the step has genuinely completed its work.
func writeDirective(ctx workflow.Context, stepID string, directive string, extraMeta map[string]any) error {
	// Write to the ResultDirective column — primary communication channel.
	if err := setResultDirective(ctx, stepID, directive); err != nil {
		return err
	}

	// Also write to status metadata for admin dashboard visibility.
	meta := map[string]any{
		DirectiveKey: directive,
	}
	for k, v := range extraMeta {
		meta[k] = v
	}

	return statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: stepID,
		Status: app.CompositeStatus{
			Status:   app.StatusSuccess,
			Metadata: meta,
		},
	})
}
