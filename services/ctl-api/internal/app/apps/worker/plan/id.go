package plan

func CreateSandboxBuildPlanWorkflowIDCallback(req *CreateSandboxBuildPlanRequest) string {
	return req.WorkflowID
}
