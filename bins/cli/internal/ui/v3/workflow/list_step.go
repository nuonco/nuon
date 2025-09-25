package workflow

import (
	"fmt"
	"math"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
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

		// this is super duper fucked up
		s := spinner.New()
		s.Spinner = spinner.Line
		now := int(math.Mod(float64(time.Now().Second()), 6))
		for _ = range now {
			s, _ = s.Update(s.Tick())
		}
		return s.View() + " " + color.Render(string(step.Status.Status))
	}

	return color.Render(string(step.Status.Status))
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
