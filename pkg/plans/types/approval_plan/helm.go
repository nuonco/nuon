package approvalplan

import (
	"github.com/tidwall/gjson"
)

type HelmApprovalPlan struct {
	PlanJSON []byte `json:"plan_json"`
}

func NewHelmApprovalPlen(planJSON []byte) *HelmApprovalPlan {
	return &HelmApprovalPlan{
		PlanJSON: planJSON,
	}
}

func (h *HelmApprovalPlan) IsNoop() (bool, error) {
	result := gjson.GetBytes(h.PlanJSON, "helm_content_diff")
	if !result.Exists() || result.IsArray() && len(result.Array()) == 0 {
		return true, nil
	}
	return false, nil
}
