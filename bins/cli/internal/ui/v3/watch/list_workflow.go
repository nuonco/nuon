package watch

import (
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type listWorkflow struct {
	workflow *models.AppWorkflow
}

func (i listWorkflow) Title() string {
	w := i.workflow
	icon := common.GetStatusIcon(w.Status.Status)
	color := styles.GetStatusStyle(w.Status.Status)
	return color.Render(icon) + " " + w.Name
}

func (i listWorkflow) Description() string {
	w := i.workflow
	finished, pending, _ := common.CalculateStepProgress(w.Steps)
	color := styles.GetStatusStyle(w.Status.Status)

	desc := color.Render(string(w.Status.Status))
	desc += fmt.Sprintf(" (%d/%d steps)", finished, finished+pending)

	return desc
}

func (i listWorkflow) FilterValue() string {
	return i.workflow.Name + " " + i.workflow.ID
}

func (i listWorkflow) Workflow() *models.AppWorkflow {
	return i.workflow
}
