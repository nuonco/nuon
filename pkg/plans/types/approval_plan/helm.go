package approvalplan

type HelmApprovalPlan struct {
	planJSON []byte `json:"plan_json"`
}

func NewHelmApprovalPlen(planJSON []byte) *HelmApprovalPlan {
	return &HelmApprovalPlan{
		planJSON: planJSON,
	}
}

func (t *HelmApprovalPlan) IsNoop() (bool, error) {
	return false, nil
}
