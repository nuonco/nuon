package approvalplan

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformApprovalPlan struct {
	planJSON []byte
}

func NewTerraformApprovalPlan(planJSON []byte) *TerraformApprovalPlan {
	return &TerraformApprovalPlan{
		planJSON: planJSON,
	}
}

// NoopPlan imeplementation for terraform is based on https://github.com/hashicorp/terraform/blob/eee744c8874f15c131651d8b34bd4860fdebcaed/internal/command/jsonformat/plan.go#L58
func terraformPlanNoop(planJSON []byte) (bool, error) {
	if len(planJSON) == 0 {
		return false, nil
	}

	var plan tfjson.Plan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return false, err
	}

	// Deferred changes are not considered current changes. If there are deferred changes,
	// infrastructure is not up to date.
	if len(plan.ResourceChanges) == 0 && len(plan.OutputChanges) == 0 && len(plan.DeferredChanges) == 0 {
		return true, nil
	}

	// A change is a no-op only when its sole action is no-op; anything else
	// (create, delete, or a replace's ["delete","create"]) is a real change.
	for _, rc := range plan.ResourceChanges {
		if rc.Change != nil && len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == tfjson.ActionNoop {
			continue
		}
		return false, nil
	}

	for _, oc := range plan.OutputChanges {
		if oc != nil && len(oc.Actions) == 1 && oc.Actions[0] == tfjson.ActionNoop {
			continue
		}
		return false, nil
	}

	for _, rd := range plan.ResourceDrift {
		if rd == nil || rd.Change == nil {
			continue
		}
		if len(rd.Change.Actions) == 1 && rd.Change.Actions[0] == tfjson.ActionNoop {
			continue
		}
		return false, nil
	}

	return true, nil
}

func (t *TerraformApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(t.planJSON)
}
