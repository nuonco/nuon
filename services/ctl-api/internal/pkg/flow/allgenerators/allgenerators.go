// Package allgenerators registers all workflow step generators into the
// generateworkflowsteps registry. Import this package in any binary that
// needs the full set of generators.
//
// This is intentionally separate from executeflow to avoid import cycles:
// executeflow is imported by signal packages that are also referenced by
// the workflow generator packages.
package allgenerators

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appworkflows "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
)

func init() {
	generateworkflowsteps.RegisterGenerators("installs", installGenerators)
	generateworkflowsteps.RegisterGenerators("apps", appGenerators)
	generateworkflowsteps.RegisterGenerators("app_branches", appGenerators)
}

func appGenerators() map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeAppConfigBuild:              appworkflows.AppConfigBuild,
		app.WorkflowTypeAppBranchesRun:              appworkflows.AppBranchRun,
		app.WorkflowTypeAppBranchesConfigRepoUpdate: appworkflows.AppBranchUpdate,
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
		app.WorkflowTypeRunbookRun:                 v2workflows.RunRunbook,
		app.WorkflowTypeAppBranchConfigUpdate:      v2workflows.AppBranchConfigUpdate,
		app.WorkflowTypeComponentEnabled:           v2workflows.ComponentEnabledSteps,
		app.WorkflowTypeComponentDisabled:          v2workflows.ComponentDisabledSteps,
	}
}
