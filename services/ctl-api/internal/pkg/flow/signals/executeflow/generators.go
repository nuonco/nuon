package executeflow

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generateworkflowsteps"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
)

func init() {
	generateworkflowsteps.RegisterGenerators("installs", installGenerators)
}

// getWorkflowStepGenerators returns the step generators for a given owner type.
// NOTE: currently only "installs" generators are registered. When additional domains
// (app-branches, etc.) adopt this signal, their generators will be added here or
// moved to a registry pattern.
func getWorkflowStepGenerators(ownerType string) map[app.WorkflowType]flow.WorkflowStepGenerator {
	switch ownerType {
	case "installs":
		return installGenerators()
	default:
		return nil
	}
}

func installGenerators() map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeManualDeploy:               v2workflows.ManualDeploySteps,
		app.WorkflowTypeDriftRun:                   v2workflows.ManualDeploySteps,
		app.WorkflowTypeDeployComponents:           v2workflows.DeployAllComponents,
		app.WorkflowTypeTeardownComponent:          v2workflows.TeardownComponent,
		app.WorkflowTypeTeardownComponents:         v2workflows.TeardownComponents,
		app.WorkflowTypeInputUpdate:                v2workflows.InputUpdate,
		app.WorkflowTypeActionWorkflowRun:          v2workflows.RunActionWorkflow,
		app.WorkflowTypeProvision:                  v2workflows.Provision,
		app.WorkflowTypeReprovision:                v2workflows.Reprovision,
		app.WorkflowTypeReprovisionSandbox:         v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox: v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDeprovision:                v2workflows.Deprovision,
		app.WorkflowTypeDeprovisionSandbox:         v2workflows.DeprovisionSandbox,
		app.WorkflowTypeSyncSecrets:                v2workflows.SyncSecrets,
	}
}
