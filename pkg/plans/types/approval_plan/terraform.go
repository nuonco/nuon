package approvalplan

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformApprovalPlan struct {
	planJSON []byte `json:"plan_json"`
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

	// Check for resource changes, output changes, and action invocations.
	noResourceChanges := len(plan.ResourceChanges) == 0
	noOutputChanges := len(plan.OutputChanges) == 0

	// Deferred changes are not considered current changes. If there are deferred changes,
	// infrastructure is not up to date.
	noDeferredChanges := len(plan.DeferredChanges) == 0

	// If all are empty, infra is up to date.
	if noResourceChanges && noOutputChanges && noDeferredChanges {
		return true, nil
	}

	// Check each resource change for no-op actions
	for _, rc := range plan.ResourceChanges {
		if rc.Change != nil && len(rc.Change.Actions) == 1 && rc.Change.Actions[0] != tfjson.ActionNoop {
			return false, nil
		}
	}

	// Check each output change for no-op actions
	for _, oc := range plan.OutputChanges {
		if oc != nil && len(oc.Actions) == 1 && oc.Actions[0] != tfjson.ActionNoop {
			return false, nil
		}
	}

	// Check each drift change for no-op actions
	for _, oc := range plan.ResourceDrift {
		if oc != nil && len(oc.Change.Actions) == 1 && oc.Change.Actions[0] != tfjson.ActionNoop {
			return false, nil
		}
	}

	return true, nil
}

func (t *TerraformApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(t.planJSON)
}
