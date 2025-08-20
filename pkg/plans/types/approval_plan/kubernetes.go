package approvalplan

type KubernetesApprovalPlan struct {
	planJSON []byte `json:"plan_json"`
}

func NewKubernetesApprovalPlan(planJSON []byte) *KubernetesApprovalPlan {
	return &KubernetesApprovalPlan{
		planJSON: planJSON,
	}
}

func (t *KubernetesApprovalPlan) IsNoop() (bool, error) {
	return false, nil
}
