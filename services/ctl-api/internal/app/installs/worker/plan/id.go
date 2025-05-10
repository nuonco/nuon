package plan

import "fmt"

// NOTE(jm): this should not be required once temporal-gen supports passing in a workflow-id via parameter
func WorkflowIDCallback(req *CreateActionRunPlanRequest) string {
	if req.WorkflowID != "" {
		return req.WorkflowID
	}

	return fmt.Sprintf("create-plan-%s", req.ActionWorkflowRunID)
}

func SandboxRunWorkflowIDCallback(req *CreateSandboxRunPlanRequest) string {
	return req.WorkflowID
}

func CreateSyncWorkflowIDCallback(req *CreateSyncPlanRequest) string {
	return req.WorkflowID
}

func CreateDeployPlanIDCallback(req *CreateDeployPlanRequest) string {
	return req.WorkflowID
}

func SyncSecretsWorkflowIDCallback(req *CreateSyncSecretsPlanRequest) string {
	return req.WorkflowID
}
