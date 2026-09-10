package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// localReadsVersion gates running the conductor's own step/group reads and the
// run-status write as local activities. Histories written before it recorded
// each one as a remote activity command, so replaying them against local
// activities is nondeterministic.
//
// Update handlers deliberately keep using the remote wrappers: a handler can run
// before the main path reaches this GetVersion, which would make the version
// marker's position in history depend on client timing.
//
// todo(sk): clean up after terminating old workflows
const localReadsVersion = "execute-flow-local-reads-v1"

// useLocalReads reports whether this execution may use the local variants.
// Repeat calls with the same change ID return the cached value, so every call
// site in a single execution agrees.
func useLocalReads(ctx workflow.Context) bool {
	return workflow.GetVersion(ctx, localReadsVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
}

func getFlowStepGroups(ctx workflow.Context, workflowID string) ([]app.WorkflowStepGroup, error) {
	if useLocalReads(ctx) {
		return workflowactivities.LocalAwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, workflowID)
	}
	return workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, workflowID)
}

func getFlowStepGroupByID(ctx workflow.Context, stepGroupID string) (*app.WorkflowStepGroup, error) {
	if useLocalReads(ctx) {
		return workflowactivities.LocalAwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, stepGroupID)
	}
	return workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroupByID(ctx, stepGroupID)
}

func getFlowSteps(ctx workflow.Context, req workflowactivities.GetFlowStepsRequest) ([]app.WorkflowStep, error) {
	if useLocalReads(ctx) {
		return workflowactivities.LocalAwaitPkgWorkflowsFlowGetFlowSteps(ctx, req)
	}
	return workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, req)
}

func getFlowStepsByFlowID(ctx workflow.Context, flowID string) ([]app.WorkflowStep, error) {
	if useLocalReads(ctx) {
		return workflowactivities.LocalAwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flowID)
	}
	return workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, flowID)
}

func updateWorkflowRunStatus(ctx workflow.Context, req workflowactivities.UpdateWorkflowRunStatusRequest) error {
	if useLocalReads(ctx) {
		return workflowactivities.LocalAwaitPkgWorkflowsFlowUpdateWorkflowRunStatus(ctx, req)
	}
	return workflowactivities.AwaitPkgWorkflowsFlowUpdateWorkflowRunStatus(ctx, req)
}
