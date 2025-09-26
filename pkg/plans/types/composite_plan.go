package plantypes

type CompositePlan struct {
	BuildPlan   *BuildPlan
	DeployPlan  *DeployPlan
	ActionPlan  *ActionWorkflowRunPlan
	SyncPlan    *SyncOCIPlan
	SandboxPlan *SandboxRunPlan
}
