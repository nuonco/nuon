package detail

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

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

	if run.TriggerType != "" {
		description += fmt.Sprintf("trigger: %s  ", run.TriggerType)
	}

	if run.CreatedBy != nil && run.CreatedBy.Email != "" {
		description += fmt.Sprintf("\nrun by: %s  ", run.CreatedBy.Email)
	}

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

func (i listRun) Run() *models.AppInstallActionWorkflowRun {
	return i.run
}
