package approvalplan

type NoopApprovalPlan struct {
	planJSON []byte
}

func NewNoopApprovalPlan(planJSON []byte) *NoopApprovalPlan {
	return &NoopApprovalPlan{
		planJSON: planJSON,
	}
}

func (t *NoopApprovalPlan) IsNoop() (bool, error) {
	return true, nil
}
