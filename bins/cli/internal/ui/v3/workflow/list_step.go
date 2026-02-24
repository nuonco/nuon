package workflow

import (
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// If we want to style the items, we'll need to write our own delegate for our custom item type

// our list step is the item we pass to the list
// it just holds a step and we implement the list item interface
// +some niecities
type listStep struct {
	step        *models.AppWorkflowStep
	spinnerView string
}

func (i listStep) Title() string {
	number := fmt.Sprintf("[%02d]", i.step.Idx)
	color := styles.GetStatusStyle(i.step.Status.Status)
	return color.Render(number) + " " + i.step.Name
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

// NOTE(fd): not in use at this time
func (i listStep) FilterValue() string {
	return i.step.Name + " " + i.step.ID
}

func (i listStep) Name() string {
	number := fmt.Sprintf("[%02d]", i.step.Idx)
	color := styles.GetStatusStyle(i.step.Status.Status)
	return color.Render(number) + " " + i.step.Name
}

// the niecities
func (i listStep) Step() *models.AppWorkflowStep {
	return i.step
}
