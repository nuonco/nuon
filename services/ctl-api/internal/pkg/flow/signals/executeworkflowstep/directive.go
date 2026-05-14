package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Step directive aliases for backward compatibility within this package.
const (
	DirectiveKey           = directive.MetadataKey
	DirectiveContinue      = directive.StepContinue
	DirectiveSkipGroup     = directive.StepSkipGroup
	DirectiveStop          = directive.StepStop
	DirectiveRetry         = directive.StepRetry
	DirectiveRetryGroup    = directive.StepRetryGroup
	DirectiveAwaitApproval = directive.StepAwaitApproval
	DirectiveAwaitRetry    = directive.StepAwaitRetry
)

// setResultDirective writes only the ResultDirective column on the step.
// Use this when the step status should NOT be changed (e.g., the step failed
// and we want to keep it as StatusError while signaling retry-group).
func setResultDirective(ctx workflow.Context, stepID string, d directive.Step) error {
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID:    stepID,
		Directive: string(d),
	})
}

// writeDirective writes a directive to both the step's ResultDirective column
// AND marks the step status as StatusSuccess with the directive in metadata.
// Use this for normal completion directives (continue, stop, skip-group, await-approval)
// where the step has genuinely completed its work.
func writeDirective(ctx workflow.Context, stepID string, d directive.Step, extraMeta map[string]any) error {
	// Write to the ResultDirective column — primary communication channel.
	if err := setResultDirective(ctx, stepID, d); err != nil {
		return err
	}

	// Also write to status metadata for admin dashboard visibility.
	meta := map[string]any{
		string(directive.MetadataKey): string(d),
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
