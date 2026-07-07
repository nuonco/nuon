package workflow

import (
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const workflowTypeDriftCheck models.AppWorkflowType = "drift_check"

func (m model) stepHasPlanDiff(step *models.AppWorkflowStep) bool {
	if m.workflow != nil && m.workflow.Type == workflowTypeDriftCheck {
		if step == nil || step.Status == nil || step.Status.Metadata == nil {
			return false
		}

		if planOnly, ok := step.Status.Metadata["plan_only"]; ok {
			return isTruthyPlanOnly(planOnly)
		}

		return false
	}

	if step == nil {
		return false
	}

	if isSystemImageSyncStep(step) {
		return false
	}

	if step.StepTargetType == "install_deploys" || step.StepTargetType == "install_deploy" {
		return true
	}

	if strings.Contains(strings.ToLower(step.Name), "plan install group") {
		return true
	}

	return strings.Contains(strings.ToLower(step.Name), "sync and plan")
}

func isSystemImageSyncStep(step *models.AppWorkflowStep) bool {
	if step == nil {
		return false
	}

	if !strings.EqualFold(strings.TrimSpace(string(step.ExecutionType)), "system") {
		return false
	}

	name := strings.ToLower(strings.TrimSpace(step.Name))

	// TODO: this name-based detection is undesirable; step kind should be explicit in API payloads.
	return strings.HasPrefix(name, "sync img_")
}

func isTruthyPlanOnly(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true")
	case float64:
		return v == 1
	case int:
		return v == 1
	}

	return false
}
