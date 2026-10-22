package plandiff

import "fmt"

// ParsedPlan is a detected plan and its structured changes.
type ParsedPlan struct {
	PlanType PlanType `json:"plan_type"`
	Plan     any      `json:"plan"`

	display string
}

// ParsePlan detects the plan type and returns the structured plan used by the formatters.
func ParsePlan(jsonStr string) (*ParsedPlan, error) {
	planType, rawPlan, err := DetectPlanType(jsonStr)
	if err != nil {
		return nil, err
	}

	switch planType {
	case PlanTypeTerraform:
		tfPlan, ok := rawPlan.(*TerraformPlan)
		if !ok {
			return nil, fmt.Errorf("failed to cast to TerraformPlan")
		}
		return &ParsedPlan{PlanType: planType, Plan: ParseTerraformPlan(tfPlan)}, nil

	case PlanTypeHelm:
		helmPlan, ok := rawPlan.(*HelmPlan)
		if !ok {
			return nil, fmt.Errorf("failed to cast to HelmPlan")
		}
		return &ParsedPlan{PlanType: planType, Plan: ParseHelmPlan(helmPlan)}, nil

	case PlanTypeKubernetes:
		k8sPlan, ok := rawPlan.(*KubernetesPlan)
		if !ok {
			return nil, fmt.Errorf("failed to cast to KubernetesPlan")
		}
		return &ParsedPlan{
			PlanType: planType,
			Plan:     ParseKubernetesPlan(k8sPlan),
			display:  k8sPlan.Plan,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported plan type: %s", planType)
	}
}

// FormatPlan detects the plan type and returns a formatted string for terminal output.
// It automatically detects whether the JSON is a Terraform, Helm, or Kubernetes plan
// and uses the appropriate parser and formatter.
func FormatPlan(jsonStr string) (string, error) {
	parsed, err := ParsePlan(jsonStr)
	if err != nil {
		return "", err
	}
	return FormatParsedPlan(parsed)
}

// FormatParsedPlan renders a plan returned by ParsePlan.
func FormatParsedPlan(parsed *ParsedPlan) (string, error) {
	if parsed == nil {
		return "", fmt.Errorf("empty plan")
	}

	switch parsed.PlanType {
	case PlanTypeTerraform:
		tfPlan, ok := parsed.Plan.(*ParsedTerraformPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to ParsedTerraformPlan")
		}
		return FormatTerraformPlan(tfPlan), nil

	case PlanTypeHelm:
		helmPlan, ok := parsed.Plan.(*ParsedHelmPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to ParsedHelmPlan")
		}
		return FormatHelmPlan(helmPlan), nil

	case PlanTypeKubernetes:
		k8sPlan, ok := parsed.Plan.(*ParsedKubernetesPlan)
		if !ok {
			return "", fmt.Errorf("failed to cast to ParsedKubernetesPlan")
		}
		return FormatKubernetesPlan(k8sPlan, parsed.display), nil

	default:
		return "", fmt.Errorf("unsupported plan type: %s", parsed.PlanType)
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
