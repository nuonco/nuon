package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// batchedStatusVersion gates the batched, local status writes. Histories written
// before it recorded the flow status, started_at and step status as three
// separate remote activity commands (and the closing status as three more), so
// replaying them against the batched local activities is nondeterministic.
//
// todo(sk): clean up after terminating old workflows
const batchedStatusVersion = "execute-workflow-step-batched-status-v1"

// useBatchedStatus reports whether this execution may use the batched local
// status writes. Repeat calls with the same change ID return the cached value,
// so every call site in a single execution agrees.
func useBatchedStatus(ctx workflow.Context) bool {
	return workflow.GetVersion(ctx, batchedStatusVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func getStep(ctx workflow.Context, stepID string) (*app.WorkflowStep, error) {
	if useBatchedStatus(ctx) {
		return activities.LocalAwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
	}
	return activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
}

func updateStepStatus(ctx workflow.Context, req statusactivities.UpdateStatusRequest) error {
	if useBatchedStatus(ctx) {
		return statusactivities.LocalAwaitPkgStatusUpdateFlowStepStatus(ctx, req)
	}
	return statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, req)
}

func updateStepFinishedAt(ctx workflow.Context, stepID string) error {
	if useBatchedStatus(ctx) {
		return activities.LocalAwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, stepID)
	}
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, stepID)
}

func updateStepResultDirective(ctx workflow.Context, req activities.UpdateFlowStepResultDirectiveRequest) error {
	if useBatchedStatus(ctx) {
		return activities.LocalAwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, req)
	}
	return activities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, req)
}
