package flows

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	WorkflowTypeProvision          app.WorkflowType = "provision"
	WorkflowTypeDeprovision        app.WorkflowType = "deprovision"
	WorkflowTypeDeprovisionSandbox app.WorkflowType = "deprovision_sandbox"

	// day-2 triggers
	WorkflowTypeManualDeploy       app.WorkflowType = "manual_deploy"
	WorkflowTypeInputUpdate        app.WorkflowType = "input_update"
	WorkflowTypeDeployComponents   app.WorkflowType = "deploy_components"
	WorkflowTypeTeardownComponent  app.WorkflowType = "teardown_component"
	WorkflowTypeTeardownComponents app.WorkflowType = "teardown_components"
	WorkflowTypeReprovisionSandbox app.WorkflowType = "reprovision_sandbox"
	WorkflowTypeActionWorkflowRun  app.WorkflowType = "action_workflow_run"

	// reprovision everything
	WorkflowTypeReprovision app.WorkflowType = "reprovision"
)

// func (i WorkflowType) PastTenseName() string {
// 	switch i {
// 	case WorkflowTypeProvision:
// 		return "Provisioned install"
// 	case WorkflowTypeReprovision:
// 		return "Reprovisioned install"
// 	case WorkflowTypeReprovisionSandbox:
// 		return "Reprovisioned sandbox"
// 	case WorkflowTypeDeprovision:
// 		return "Deprovisioned install"
// 	case WorkflowTypeManualDeploy:
// 		return "Deployed to install"
// 	case WorkflowTypeInputUpdate:
// 		return "Updated Input"
// 	case WorkflowTypeTeardownComponents:
// 		return "Tore down all components"
// 	case WorkflowTypeDeployComponents:
// 		return "Deployed all components"
// 	default:
// 	}

// 	return ""
// }

// func (i WorkflowType) Name() string {
// 	switch i {
// 	case WorkflowTypeProvision:
// 		return "Provisioning install"
// 	case WorkflowTypeReprovision:
// 		return "Reprovisioning install"
// 	case WorkflowTypeDeprovision:
// 		return "Deprovisioning install"
// 	case WorkflowTypeManualDeploy:
// 		return "Deploying to install"
// 	case WorkflowTypeInputUpdate:
// 		return "Input Update"
// 	case WorkflowTypeTeardownComponents:
// 		return "Tearing down all components"
// 	case WorkflowTypeDeployComponents:
// 		return "Deploying all components"
// 	case WorkflowTypeReprovisionSandbox:
// 		return "Reprovisioning sandbox"
// 	default:
// 	}

// 	return ""
// }

// func (i WorkflowType) Description() string {
// 	switch i {
// 	case WorkflowTypeProvision:
// 		return "Creates a runner stack, waits for it to be applied and then provisions the sandbox and deploys all components."
// 	case WorkflowTypeReprovision:
// 		return "Creates a new runner stack, waits for it to be applied and then reprovisions the sandbox and deploys all components."
// 	case WorkflowTypeReprovisionSandbox:
// 		return "Reprovisions the sandbox and redeploys everything on top of it."
// 	case WorkflowTypeDeprovision:
// 		return "Deprovisions all components, deprovisions the sandbox and then waits for the cloudformation stack to be deleted."
// 	case WorkflowTypeManualDeploy:
// 		return "Deploys a single component."
// 	case WorkflowTypeInputUpdate:
// 		return "Depending on which input was changed, will reprovision the sandbox and deploy one or all components."
// 	case WorkflowTypeDeployComponents:
// 		return "Deploy all components in the order of their dependencies."
// 	case WorkflowTypeTeardownComponents:
// 		return "Teardown components in the reverse order of their dependencies."
// 	}

// 	return "unknown"
// }
