package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/flowutil"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// Re-export directive constants for convenience within this package.
const (
	DirectiveKey           = flowutil.DirectiveKey
	DirectiveContinue      = flowutil.DirectiveContinue
	DirectiveSkipGroup     = flowutil.DirectiveSkipGroup
	DirectiveStop          = flowutil.DirectiveStop
	DirectiveRetry         = flowutil.DirectiveRetry
	DirectiveAwaitApproval = flowutil.DirectiveAwaitApproval
)

// writeDirective writes a directive to the step's status metadata so the parent
// conductor can read it after the step signal completes.
func writeDirective(ctx workflow.Context, stepID string, directive string, extraMeta map[string]any) error {
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
