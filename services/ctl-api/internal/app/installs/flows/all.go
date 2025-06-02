package flows

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	FlowTypeProvision          app.FlowType = "provision"
	FlowTypeDeprovision        app.FlowType = "deprovision"
	FlowTypeDeprovisionSandbox app.FlowType = "deprovision_sandbox"

	// day-2 triggers
	FlowTypeManualDeploy       app.FlowType = "manual_deploy"
	FlowTypeInputUpdate        app.FlowType = "input_update"
	FlowTypeDeployComponents   app.FlowType = "deploy_components"
	FlowTypeTeardownComponent  app.FlowType = "teardown_component"
	FlowTypeTeardownComponents app.FlowType = "teardown_components"
	FlowTypeReprovisionSandbox app.FlowType = "reprovision_sandbox"
	FlowTypeActionWorkflowRun  app.FlowType = "action_workflow_run"

	// reprovision everything
	FlowTypeReprovision app.FlowType = "reprovision"
)

// func (i FlowType) PastTenseName() string {
// 	switch i {
// 	case FlowTypeProvision:
// 		return "Provisioned install"
// 	case FlowTypeReprovision:
// 		return "Reprovisioned install"
// 	case FlowTypeReprovisionSandbox:
// 		return "Reprovisioned sandbox"
// 	case FlowTypeDeprovision:
// 		return "Deprovisioned install"
// 	case FlowTypeManualDeploy:
// 		return "Deployed to install"
// 	case FlowTypeInputUpdate:
// 		return "Updated Input"
// 	case FlowTypeTeardownComponents:
// 		return "Tore down all components"
// 	case FlowTypeDeployComponents:
// 		return "Deployed all components"
// 	default:
// 	}

// 	return ""
// }

// func (i FlowType) Name() string {
// 	switch i {
// 	case FlowTypeProvision:
// 		return "Provisioning install"
// 	case FlowTypeReprovision:
// 		return "Reprovisioning install"
// 	case FlowTypeDeprovision:
// 		return "Deprovisioning install"
// 	case FlowTypeManualDeploy:
// 		return "Deploying to install"
// 	case FlowTypeInputUpdate:
// 		return "Input Update"
// 	case FlowTypeTeardownComponents:
// 		return "Tearing down all components"
// 	case FlowTypeDeployComponents:
// 		return "Deploying all components"
// 	case FlowTypeReprovisionSandbox:
// 		return "Reprovisioning sandbox"
// 	default:
// 	}

// 	return ""
// }

// func (i FlowType) Description() string {
// 	switch i {
// 	case FlowTypeProvision:
// 		return "Creates a runner stack, waits for it to be applied and then provisions the sandbox and deploys all components."
// 	case FlowTypeReprovision:
// 		return "Creates a new runner stack, waits for it to be applied and then reprovisions the sandbox and deploys all components."
// 	case FlowTypeReprovisionSandbox:
// 		return "Reprovisions the sandbox and redeploys everything on top of it."
// 	case FlowTypeDeprovision:
// 		return "Deprovisions all components, deprovisions the sandbox and then waits for the cloudformation stack to be deleted."
// 	case FlowTypeManualDeploy:
// 		return "Deploys a single component."
// 	case FlowTypeInputUpdate:
// 		return "Depending on which input was changed, will reprovision the sandbox and deploy one or all components."
// 	case FlowTypeDeployComponents:
// 		return "Deploy all components in the order of their dependencies."
// 	case FlowTypeTeardownComponents:
// 		return "Teardown components in the reverse order of their dependencies."
// 	}

// 	return "unknown"
// }
