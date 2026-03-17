package plan

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
)

// TerraformPlan wraps the Terraform JSON plan in a format
// expected by OPA policies prevalent in the ecosystem.
type TerraformPlan struct {
	Plan tfjson.Plan `json:"plan"`
}

// ParseTerraformPlan parses the given Terraform plan JSON into a TerraformPlan structure.
func ParseTerraformPlan(planJSON []byte) (*TerraformPlan, error) {
	var plan tfjson.Plan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return nil, err
	}

	return &TerraformPlan{
		Plan: plan,
	}, nil
}
