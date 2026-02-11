package plandiff

import "fmt"

// FormatPlan detects the plan type and returns a formatted string for terminal output.
// It automatically detects whether the JSON is a Terraform, Helm, or Kubernetes plan
// and uses the appropriate parser and formatter.
func FormatPlan(jsonStr string) (string, error) {
	planType, rawPlan, err := DetectPlanType(jsonStr)
	if err != nil {
		return "", err
	}

	switch planType {
	case PlanTypeTerraform:
		tfPlan, ok := rawPlan.(*TerraformPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to TerraformPlan")
		}
		parsed := ParseTerraformPlan(tfPlan)
		return FormatTerraformPlan(parsed), nil

	case PlanTypeHelm:
		helmPlan, ok := rawPlan.(*HelmPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to HelmPlan")
		}
		parsed := ParseHelmPlan(helmPlan)
		return FormatHelmPlan(parsed), nil

	case PlanTypeKubernetes:
		k8sPlan, ok := rawPlan.(*KubernetesPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to KubernetesPlan")
		}
		parsed := ParseKubernetesPlan(k8sPlan)
		return FormatKubernetesPlan(parsed, k8sPlan.Plan), nil

	default:
		return "", fmt.Errorf("unknown plan type: %s", planType)
	}
}

// HasChanges detects the plan type and returns true if the plan has any meaningful changes.
// This can be used to determine if approval is needed or if the plan is a no-op.
func HasChanges(jsonStr string) (bool, error) {
	planType, rawPlan, err := DetectPlanType(jsonStr)
	if err != nil {
		return false, err
	}

	switch planType {
	case PlanTypeTerraform:
		tfPlan, ok := rawPlan.(*TerraformPlan)
		if !ok {
			return false, fmt.Errorf("failed to cast to TerraformPlan")
		}
		parsed := ParseTerraformPlan(tfPlan)
		return HasTerraformChanges(parsed), nil

	case PlanTypeHelm:
		helmPlan, ok := rawPlan.(*HelmPlan)
		if !ok {
			return false, fmt.Errorf("failed to cast to HelmPlan")
		}
		parsed := ParseHelmPlan(helmPlan)
		return HasHelmChanges(parsed), nil

	case PlanTypeKubernetes:
		k8sPlan, ok := rawPlan.(*KubernetesPlan)
		if !ok {
			return false, fmt.Errorf("failed to cast to KubernetesPlan")
		}
		parsed := ParseKubernetesPlan(k8sPlan)
		return HasKubernetesChanges(parsed), nil

	default:
		return false, fmt.Errorf("unknown plan type: %s", planType)
	}
}

// GetPlanType returns the detected plan type for a JSON string.
// This is useful when you need to know the type before formatting.
func GetPlanType(jsonStr string) (PlanType, error) {
	planType, _, err := DetectPlanType(jsonStr)
	return planType, err
}
