package app

import "testing"

func TestWorkflowTypeRequiresLiveInstallRunner(t *testing.T) {
	noLive := []WorkflowType{
		WorkflowTypeProvision,
		WorkflowTypeReprovision,
		WorkflowTypeReprovisionStack,
		WorkflowTypeAppBranchesRun,
		WorkflowTypeAppBranchesConfigRepoUpdate,
		WorkflowTypeAppBranchesComponentRepoUpdate,
		WorkflowTypeAppInstallSync,
		WorkflowTypeAppConfigBuild,
	}
	for _, wt := range noLive {
		if wt.RequiresLiveInstallRunner() {
			t.Errorf("%s should not require a live install runner", wt)
		}
	}

	live := []WorkflowType{
		WorkflowTypeDeprovision,
		WorkflowTypeDeprovisionSandbox,
		WorkflowTypeManualDeploy,
		WorkflowTypeInputUpdate,
		WorkflowTypeDeployComponents,
		WorkflowTypeTeardownComponent,
		WorkflowTypeTeardownComponents,
		WorkflowTypeReprovisionSandbox,
		WorkflowTypeDriftRunReprovisionSandbox,
		WorkflowTypeActionWorkflowRun,
		WorkflowTypeSyncSecrets,
		WorkflowTypeDriftRun,
		WorkflowTypeRunbookRun,
		WorkflowTypeComponentEnabled,
		WorkflowTypeComponentDisabled,
		WorkflowTypeAppBranchConfigUpdate,
		WorkflowTypeRecoverHelmRelease,
	}
	for _, wt := range live {
		if !wt.RequiresLiveInstallRunner() {
			t.Errorf("%s should require a live install runner", wt)
		}
	}

	if WorkflowType("some_future_workflow").RequiresLiveInstallRunner() {
		t.Error("unknown workflow types should not require a live install runner")
	}
}
