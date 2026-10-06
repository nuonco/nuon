package appbundle

import (
	"encoding/json"
	"fmt"
)

func clearTerraformPlanArtifacts(plan json.RawMessage) (json.RawMessage, error) {
	var composite map[string]any
	if err := json.Unmarshal(plan, &composite); err != nil {
		return nil, err
	}
	deploy, ok := composite["deploy_plan"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("composite plan has no deploy_plan")
	}
	deploy["apply_plan_contents"] = ""
	deploy["apply_plan_display"] = ""
	terraformPlan, ok := deploy["terraform"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("deploy plan has no terraform plan")
	}
	terraformPlan["plan_json"] = nil
	return json.Marshal(composite)
}
