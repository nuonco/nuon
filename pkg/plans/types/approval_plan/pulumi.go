package approvalplan

import (
	"encoding/json"
)

type PulumiApprovalPlan struct {
	planJSON []byte
}

func NewPulumiApprovalPlan(planJSON []byte) *PulumiApprovalPlan {
	return &PulumiApprovalPlan{
		planJSON: planJSON,
	}
}

func (p *PulumiApprovalPlan) IsNoop() (bool, error) {
	if len(p.planJSON) == 0 {
		return false, nil
	}

	var preview struct {
		ChangeSummary map[string]int `json:"change_summary"`
	}
	if err := json.Unmarshal(p.planJSON, &preview); err != nil {
		return false, err
	}

	for action, count := range preview.ChangeSummary {
		if action == "same" {
			continue
		}
		if count > 0 {
			return false, nil
		}
	}

	return true, nil
}
