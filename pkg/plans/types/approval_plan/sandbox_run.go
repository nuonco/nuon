package approvalplan

type SandboxRunApprovalPlan struct {
	// in case of sandbox it'll be terraform style plan
	PlanJSON []byte `json:"plan_json"`
}

func NewSandboxRunApprovalPlan(planJSON []byte) *SandboxRunApprovalPlan {
	return &SandboxRunApprovalPlan{
		PlanJSON: planJSON,
	}
}

func (s *SandboxRunApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(s.PlanJSON)
}
