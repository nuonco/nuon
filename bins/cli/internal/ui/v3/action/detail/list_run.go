package detail

import (
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// our list run is the item we pass to the list
// it just holds a run and we implement the list item interface
type listRun struct {
	run  *models.AppInstallActionWorkflowRun
	name string
}

func (i listRun) Title() string {
	run := i.run
	statusStyle := styles.GetStatusStyle(run.StatusV2.Status)

	return statusStyle.Render(fmt.Sprintf("[%s] ", run.StatusV2.Status)) + i.name
}

func (i listRun) Description() string {
	run := i.run
	description := ""

	// trigger type
	if run.TriggerType != "" {
		description += fmt.Sprintf("trigger: %s  ", run.TriggerType)
	}

	// created by
	if run.CreatedBy != nil && run.CreatedBy.Email != "" {
		description += fmt.Sprintf("\nrun by: %s  ", run.CreatedBy.Email)
	}

	// humanized time
	if run.CreatedAt != "" {
		description += run.CreatedAt
	}

	return description
}

func (i listRun) FilterValue() string {
	run := i.run
	filterStr := run.ID
	if run.InstallActionWorkflow != nil && run.InstallActionWorkflow.ActionWorkflow != nil {
		filterStr += " " + run.InstallActionWorkflow.ActionWorkflow.Name
	}
	if run.CreatedBy != nil {
		filterStr += " " + run.CreatedBy.Email
	}
	return filterStr
}

// the niecities
func (i listRun) Run() *models.AppInstallActionWorkflowRun {
	return i.run
}
