package steps

import (
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// stepItem represents a step with both config and run data
type stepItem struct {
	configStep *models.AppActionWorkflowStepConfig
	runStep    *models.AppInstallActionWorkflowRunStep
}

// getName returns the name of the step
func (s stepItem) getName() string {
	if s.configStep != nil && s.configStep.Name != "" {
		return s.configStep.Name
	}
	return styles.TextDim.Render("Unnamed Step")
}

// getStatus returns the status of the step
func (s stepItem) getStatus() string {
	if s.runStep == nil {
		return "pending"
	}
	return string(s.runStep.Status)
}

// getExecutionDuration returns the execution duration of the step
func (s stepItem) getExecutionDuration() string {
	if s.runStep == nil || s.runStep.ExecutionDuration == 0 {
		return ""
	}
	duration := common.HumanizeNSDuration(s.runStep.ExecutionDuration)
	return duration
}
