package workflow

import (
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// +some niecities
type listStep struct {
	step        *models.AppWorkflowStep
	spinnerView string
}

func (i listStep) Title() string {
	color := styles.GetStatusStyle(i.step.Status.Status)
	return color.Render(i.statusIcon()) + " " + i.step.Name
}

func (i listStep) statusIcon() string {
	if common.IsInProgressStatus(i.step.Status.Status) && i.spinnerView != "" {
		return i.spinnerView
	}
	return common.GetStatusIcon(i.step.Status.Status)
}

func (i listStep) Description() string {
	step := i.step
	if generics.SliceContains(step.Status.Status, terminalStatuses) {
		return fmt.Sprintf("executed in %s", common.HumanizeNSDuration(i.step.ExecutionTime))
	}

	color := styles.GetStatusStyle(step.Status.Status)
	if i.step.Status.Status == models.AppStatusInDashProgress {
		return i.spinnerView + " " + color.Render(string(step.Status.Status))
	}

	return color.Render(string(step.Status.Status))
}

func (i listStep) FilterValue() string {
	return i.step.Name + " " + i.step.ID
}

func (i listStep) Name() string {
	return i.Title()
}

func (i listStep) Step() *models.AppWorkflowStep {
	return i.step
}
