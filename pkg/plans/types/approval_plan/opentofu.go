package approvalplan

// OpenTofu emits the same JSON plan schema as Terraform, so noop detection
// reuses the terraform implementation.
type OpentofuApprovalPlan struct {
	planJSON []byte
}

func NewOpentofuApprovalPlan(planJSON []byte) *OpentofuApprovalPlan {
	return &OpentofuApprovalPlan{
		planJSON: planJSON,
	}
}

func (t *OpentofuApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(t.planJSON)
}
