package app

import "testing"

func TestWorkflowTypeRequiresInstallRunner(t *testing.T) {
	if WorkflowTypeReprovisionStack.RequiresInstallRunner() {
		t.Error("reprovision_stack should not require the install runner")
	}

	needsRunner := []WorkflowType{
		WorkflowTypeProvision,
		WorkflowTypeReprovision,
		WorkflowTypeReprovisionSandbox,
		WorkflowTypeDriftRunReprovisionSandbox,
		WorkflowTypeDeprovision,
		WorkflowTypeDeprovisionSandbox,
		WorkflowTypeManualDeploy,
		WorkflowTypeDriftRun,
		WorkflowTypeDeployComponents,
		WorkflowTypeTeardownComponent,
		WorkflowTypeTeardownComponents,
		WorkflowTypeInputUpdate,
		WorkflowTypeActionWorkflowRun,
		WorkflowTypeSyncSecrets,
		WorkflowTypeRunbookRun,
		WorkflowTypeComponentEnabled,
		WorkflowTypeComponentDisabled,
		WorkflowTypeRecoverHelmRelease,
		WorkflowTypeAppBranchConfigUpdate,
	}
	for _, wt := range needsRunner {
		if !wt.RequiresInstallRunner() {
			t.Errorf("%s should require the install runner", wt)
		}
	}

	if !WorkflowType("some_future_workflow").RequiresInstallRunner() {
		t.Error("unknown workflow types should require the install runner")
	}
}
