package approvalplan

type NoopApprovalPlan struct {
	planJSON []byte `json:"plan_json"`
}

func NewNoopApprovalPlan(planJSON []byte) *NoopApprovalPlan {
	return &NoopApprovalPlan{
		planJSON: planJSON,
	}
}

func (t *NoopApprovalPlan) IsNoop() (bool, error) {
	return true, nil
}
