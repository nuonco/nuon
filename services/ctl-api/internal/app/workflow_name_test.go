package app

import "testing"

// every type declared in workflow.go — add new ones here so the titles below
// stay mandatory rather than silently falling back to the lowercased type
var allWorkflowTypes = []WorkflowType{
	WorkflowTypeProvision,
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
	WorkflowTypeAppBranchesRun,
	WorkflowTypeAppBranchesConfigRepoUpdate,
	WorkflowTypeAppBranchesComponentRepoUpdate,
	WorkflowTypeAppBranchConfigUpdate,
	WorkflowTypeReprovision,
	WorkflowTypeReprovisionStack,
	WorkflowTypeAppConfigBuild,
	WorkflowTypeRunbookRun,
	WorkflowTypeComponentEnabled,
	WorkflowTypeComponentDisabled,
}

func TestWorkflowTypeTitlesAreExhaustive(t *testing.T) {
	for _, wt := range allWorkflowTypes {
		// app_branches_manual_update is titled from its metadata instead
		if wt == WorkflowTypeAppBranchesRun {
			continue
		}
		if wt.Name() == "" {
			t.Errorf("%s has no Name()", wt)
		}
		if wt.PastTenseName() == "" {
			t.Errorf("%s has no PastTenseName()", wt)
		}
	}
}

func TestComputeNameNeverFallsBackToRawType(t *testing.T) {
	for _, wt := range allWorkflowTypes {
		name := (&Workflow{Type: wt}).ComputeName()
		if name == "" {
			t.Errorf("%s computed an empty name", wt)
		}
		if name == string(wt) {
			t.Errorf("%s computed the raw type as its name", wt)
		}
	}
}
