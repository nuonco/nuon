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
		return !h.releaseNeedsDeploy(), nil
	}
	return false, nil
}

// A timed-out attempt records the new manifest on a pending/failed release, so the
// retry diffs clean even though nothing rolled out. A missing status (old runner, or
// no previous release info) is treated as needing deploy so an empty diff never
// silently skips an apply whose release state is unknown.
func (h *HelmApprovalPlan) releaseNeedsDeploy() bool {
	status := gjson.GetBytes(h.PlanJSON, "helm_release_status")
	if !status.Exists() || status.String() == "" {
		return true
	}
	return status.String() != helmStatusDeployed
}

const helmStatusDeployed = "deployed"
