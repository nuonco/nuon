package approvalplan

type NoopApprovalPlan struct {
	PlanJSON []byte `json:"plan_json"`
}

func NewNoopApprovalPlan(planJSON []byte) *NoopApprovalPlan {
	return &NoopApprovalPlan{
		PlanJSON: planJSON,
	}
}

func (t *NoopApprovalPlan) IsNoop() (bool, error) {
	return true, nil
}
