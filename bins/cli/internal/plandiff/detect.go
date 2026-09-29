package plandiff

import (
	"encoding/json"
	"fmt"
)

func DetectPlanType(jsonStr string) (PlanType, any, error) {
	if jsonStr == "" {
		return PlanTypeUnknown, nil, fmt.Errorf("empty plan string")
	}

	extractedPlan := extractPlanFromWrapper(jsonStr)

	if planType, plan, err := tryTerraform(extractedPlan); err == nil {
		return planType, plan, nil
	}

	if planType, plan, err := tryHelm(extractedPlan); err == nil {
		return planType, plan, nil
	}

	if planType, plan, err := tryKubernetes(extractedPlan); err == nil {
		return planType, plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("unable to detect plan type")
}

func extractPlanFromWrapper(jsonStr string) string {
	var wrapper RunnerJobPlanWrapper
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return jsonStr
	}

	if wrapper.SandboxMode != nil {
		if wrapper.SandboxMode.KubernetesManifest != nil && wrapper.SandboxMode.KubernetesManifest.PlanContents != "" {
			return wrapper.SandboxMode.KubernetesManifest.PlanContents
		}
		if wrapper.SandboxMode.Helm != nil && wrapper.SandboxMode.Helm.PlanContents != "" {
			return wrapper.SandboxMode.Helm.PlanContents
		}
		if wrapper.SandboxMode.Terraform != nil && wrapper.SandboxMode.Terraform.PlanDisplayContents != "" {
			return wrapper.SandboxMode.Terraform.PlanDisplayContents
		}
	}

	if wrapper.KubernetesManifest != nil && wrapper.KubernetesManifest.PlanContents != "" {
		return wrapper.KubernetesManifest.PlanContents
	}
	if wrapper.Helm != nil && wrapper.Helm.PlanContents != "" {
		return wrapper.Helm.PlanContents
	}
	if wrapper.Terraform != nil && wrapper.Terraform.PlanDisplayContents != "" {
		return wrapper.Terraform.PlanDisplayContents
	}

	if wrapper.ApplyPlanContents != "" {
		return wrapper.ApplyPlanContents
	}

	return jsonStr
}

func tryTerraform(jsonStr string) (PlanType, *TerraformPlan, error) {
	var plan TerraformPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	if plan.ResourceChanges != nil {
		return PlanTypeTerraform, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a terraform plan: missing resource_changes")
}

func tryHelm(jsonStr string) (PlanType, *HelmPlan, error) {
	var plan HelmPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	if plan.HelmContentDiff != nil {
		return PlanTypeHelm, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a helm plan: missing helm_content_diff")
}

func tryKubernetes(jsonStr string) (PlanType, *KubernetesPlan, error) {
	var plan KubernetesPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return PlanTypeUnknown, nil, err
	}

	if plan.K8sContentDiff != nil {
		return PlanTypeKubernetes, &plan, nil
	}

	return PlanTypeUnknown, nil, fmt.Errorf("not a kubernetes plan: missing k8s_content_diff")
}

func MustParseTerraform(jsonStr string) *TerraformPlan {
	_, plan, err := tryTerraform(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}

func MustParseHelm(jsonStr string) *HelmPlan {
	_, plan, err := tryHelm(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}

func MustParseKubernetes(jsonStr string) *KubernetesPlan {
	_, plan, err := tryKubernetes(jsonStr)
	if err != nil {
		panic(err)
	}
	return plan
}
