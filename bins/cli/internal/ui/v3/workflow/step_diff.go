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

	if step.StepTargetType == "install_deploys" || step.StepTargetType == "install_deploy" {
		return true
	}

	return strings.Contains(strings.ToLower(step.Name), "sync and plan")
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
