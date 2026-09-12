package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// localReadsVersion gates running the group's own reads and status writes as
// local activities. Histories written before it recorded each one as a remote
// activity command, so replaying them against local activities is
// nondeterministic.
//
// todo(sk): clean up after terminating old workflows
const localReadsVersion = "execute-workflow-step-group-local-reads-v1"

// useLocalReads reports whether this execution may use the local variants.
// Repeat calls with the same change ID return the cached value, so every call
// site in a single execution agrees.
func useLocalReads(ctx workflow.Context) bool {
	return workflow.GetVersion(ctx, localReadsVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func getFlowSteps(ctx workflow.Context, flowID string) ([]app.WorkflowStep, error) {
	req := activities.GetFlowStepsRequest{FlowID: flowID}
	if useLocalReads(ctx) {
		return activities.LocalAwaitPkgWorkflowsFlowGetFlowSteps(ctx, req)
	}
	return activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, req)
}

func getStep(ctx workflow.Context, stepID string) (*app.WorkflowStep, error) {
	if useLocalReads(ctx) {
		return activities.LocalAwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
	}
	return activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, stepID)
}

func updateStepStatus(ctx workflow.Context, req statusactivities.UpdateStatusRequest) error {
	if useLocalReads(ctx) {
		return statusactivities.LocalAwaitPkgStatusUpdateFlowStepStatus(ctx, req)
	}
	return statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, req)
}

func updateGroupStatus(ctx workflow.Context, req statusactivities.UpdateStatusRequest) error {
	if useLocalReads(ctx) {
		return statusactivities.LocalAwaitPkgStatusUpdateFlowStepGroupStatus(ctx, req)
	}
	return statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, req)
}
