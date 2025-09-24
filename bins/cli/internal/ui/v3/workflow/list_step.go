package workflow

import (
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// If we want to style the items, we'll need to write our own delegate for our custom item type

// our list step is the item we pass to the list
// it just holds a step and we implement the list item interface
// +some niecities
type listStep struct {
	step *models.AppWorkflowStep
}

func (i listStep) Title() string {
	number := fmt.Sprintf("[%02d]", i.step.Idx)
	title := number + " " + i.step.Name
	return title
}

var terminalStatuses = []models.AppStatus{
	models.AppStatusCancelled,
	models.AppStatusError,
	models.AppStatusSuccess,
	// models.AppStatusFailed

}

func (i listStep) Description() string {
	step := i.step
	if generics.SliceContains(step.Status.Status, terminalStatuses) {
		return fmt.Sprintf("executed in %s", common.HumanizeNSDuration(i.step.ExecutionTime))
	}
	return string(step.Status.Status)
}

func (i listStep) FilterValue() string {
	return i.step.Name
}

func (i listStep) Name() string {
	return i.step.ID
}

// the niecities
func (i listStep) Step() *models.AppWorkflowStep {
	return i.step
}
