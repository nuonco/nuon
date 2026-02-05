package plandiff

import (
	"encoding/json"
	"fmt"
)

// DetectPlanType attempts to detect the plan type from a JSON string
// and returns the plan type along with the parsed plan structure.
// It first tries to unwrap the runner job API response, then detects the plan type.
func DetectPlanType(jsonStr string) (PlanType, any, error) {
	if jsonStr == "" {
		return PlanTypeUnknown, nil, fmt.Errorf("empty plan string")
	}

	// First, try to extract the actual plan from the runner job wrapper
	extractedPlan := extractPlanFromWrapper(jsonStr)

	// Try Terraform first (has resource_changes)
	if planType, plan, err := tryTerraform(extractedPlan); err == nil {
		return planType, plan, nil
	}

	// Try Helm (has helm_content_diff)
	if planType, plan, err := tryHelm(extractedPlan); err == nil {
		return planType, plan, nil
	}

	// Try Kubernetes (has k8s_content_diff)
	if planType, plan, err := tryKubernetes(extractedPlan); err == nil {
		return planType, plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("unable to detect plan type")
}

// extractPlanFromWrapper attempts to extract the nested plan from a runner job API response.
// If the input is already a raw plan, it returns the original string.
func extractPlanFromWrapper(jsonStr string) string {
	var wrapper RunnerJobPlanWrapper
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return jsonStr
	}

	// Check sandbox_mode first (most common for sandbox deployments)
	if wrapper.SandboxMode != nil {
		if wrapper.SandboxMode.KubernetesManifest != nil && wrapper.SandboxMode.KubernetesManifest.PlanContents != "" {
			return wrapper.SandboxMode.KubernetesManifest.PlanContents
		}
		if wrapper.SandboxMode.Helm != nil && wrapper.SandboxMode.Helm.PlanContents != "" {
			return wrapper.SandboxMode.Helm.PlanContents
		}
		// Terraform uses plan_display_contents for the JSON plan (plan_contents is binary)
		if wrapper.SandboxMode.Terraform != nil && wrapper.SandboxMode.Terraform.PlanDisplayContents != "" {
			return wrapper.SandboxMode.Terraform.PlanDisplayContents
		}
	}

	// Check top-level component fields (non-sandbox mode)
	if wrapper.KubernetesManifest != nil && wrapper.KubernetesManifest.PlanContents != "" {
		return wrapper.KubernetesManifest.PlanContents
	}
	if wrapper.Helm != nil && wrapper.Helm.PlanContents != "" {
		return wrapper.Helm.PlanContents
	}
	// Terraform uses plan_display_contents for the JSON plan (plan_contents is binary)
	if wrapper.Terraform != nil && wrapper.Terraform.PlanDisplayContents != "" {
		return wrapper.Terraform.PlanDisplayContents
	}

	// Check apply_plan_contents (Terraform non-sandbox)
	if wrapper.ApplyPlanContents != "" {
		return wrapper.ApplyPlanContents
	}

	// No wrapper detected, return original
	return jsonStr
}

// tryTerraform attempts to parse as a Terraform plan
func tryTerraform(jsonStr string) (PlanType, *TerraformPlan, error) {
	var plan TerraformPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	// Terraform plans must have resource_changes array
	if plan.ResourceChanges != nil {
		return PlanTypeTerraform, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a terraform plan: missing resource_changes")
}

// tryHelm attempts to parse as a Helm plan
func tryHelm(jsonStr string) (PlanType, *HelmPlan, error) {
	var plan HelmPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	// Helm plans must have helm_content_diff array
	if plan.HelmContentDiff != nil {
		return PlanTypeHelm, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a helm plan: missing helm_content_diff")
}

// tryKubernetes attempts to parse as a Kubernetes plan
func tryKubernetes(jsonStr string) (PlanType, *KubernetesPlan, error) {
	var plan KubernetesPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	// Kubernetes plans must have k8s_content_diff array
	if plan.K8sContentDiff != nil {
		return PlanTypeKubernetes, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a kubernetes plan: missing k8s_content_diff")
}

// MustParseTerraform parses a JSON string as a Terraform plan or panics
func MustParseTerraform(jsonStr string) *TerraformPlan {
	_, plan, err := tryTerraform(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}

// MustParseHelm parses a JSON string as a Helm plan or panics
func MustParseHelm(jsonStr string) *HelmPlan {
	_, plan, err := tryHelm(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}

// MustParseKubernetes parses a JSON string as a Kubernetes plan or panics
func MustParseKubernetes(jsonStr string) *KubernetesPlan {
	_, plan, err := tryKubernetes(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}
