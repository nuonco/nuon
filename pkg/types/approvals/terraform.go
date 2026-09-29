package plan

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformPlan struct {
	Plan tfjson.Plan `json:"plan"`
}

func ParseTerraformPlan(planJSON []byte) (*TerraformPlan, error) {
	var plan tfjson.Plan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return nil, err
	}

	return &TerraformPlan{
		Plan: plan,
	}, nil
}
