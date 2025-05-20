package plan

func CreateComponentBuildWorkflowIDCallback(req *CreateComponentBuildPlanRequest) string {
	return req.WorkflowID
}
