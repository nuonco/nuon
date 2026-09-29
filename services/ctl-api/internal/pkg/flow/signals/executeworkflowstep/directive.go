package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

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

func setResultDirective(ctx workflow.Context, stepID string, d directive.Step) error {
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, activities.UpdateFlowStepResultDirectiveRequest{
		StepID:    stepID,
		Directive: string(d),
	})
}

func writeDirective(ctx workflow.Context, stepID string, d directive.Step, extraMeta map[string]any) error {
	if err := setResultDirective(ctx, stepID, d); err != nil {
		return err
	}

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
