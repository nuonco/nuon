package preflight

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
)

func Run(ctx workflow.Context, flw *app.Workflow, generated *app.GenerateStepsResult) (*flow.WorkflowPreflightResult, error) {
	result, err := activities.AwaitInstallPreflight(ctx, installPreflightRequest(flw, generated))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check install workflow prerequisites")
	}
	return &flow.WorkflowPreflightResult{Findings: result.Findings}, nil
}

func installPreflightRequest(flw *app.Workflow, generated *app.GenerateStepsResult) activities.InstallPreflightRequest {
	return activities.InstallPreflightRequest{
		FlowID:                 flw.ID,
		InstallID:              flw.OwnerID,
		DesiredAppConfigID:     generics.FromPtrStr(flw.Metadata["new_app_config_id"]),
		CheckStackOutdated:     stackOutdatedCheckApplies(flw.Type) && !regeneratesStack(generated),
		PlannedComponentBuilds: plannedComponentBuilds(generated),
	}
}

func stackOutdatedCheckApplies(workflowType app.WorkflowType) bool {
	return workflowType != app.WorkflowTypeDeprovision && workflowType != app.WorkflowTypeDeprovisionSandbox
}

func regeneratesStack(generated *app.GenerateStepsResult) bool {
	if generated == nil {
		return false
	}
	for _, step := range generated.Steps {
		if step == nil || step.ExecutionType == app.WorkflowStepExecutionTypeSkipped || step.QueueSignal == nil {
			continue
		}
		if _, ok := step.QueueSignal.Signal.(*generateinstallstackversion.Signal); ok {
			return true
		}
	}
	return false
}
