package approvalplan

type SandboxRunApprovalPlan struct {
	// in case of sandbox it'll be terraform style plan
	planJSON []byte `json:"plan_json"`
}

func NewSandboxRunApprovalPlan(planJSON []byte) *SandboxRunApprovalPlan {
	return &SandboxRunApprovalPlan{
		planJSON: planJSON,
	}
}

func (s *SandboxRunApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(s.planJSON)
}
