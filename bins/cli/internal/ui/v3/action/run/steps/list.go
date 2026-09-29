package steps

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type stepItem struct {
	configStep *models.AppActionWorkflowStepConfig
	runStep    *models.AppInstallActionWorkflowRunStep
}

func (s stepItem) getName() string {
	if s.configStep != nil && s.configStep.Name != "" {
		return s.configStep.Name
	}
	return styles.TextDim.Render("Unnamed Step")
}

func (s stepItem) getStatus() string {
	if s.runStep == nil {
		return "pending"
	}
	return string(s.runStep.Status)
}

func (s stepItem) getExecutionDuration() string {
	if s.runStep == nil || s.runStep.ExecutionDuration == 0 {
		return ""
	}
	duration := common.HumanizeNSDuration(s.runStep.ExecutionDuration)
	return duration
}
