package run

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type installActionWorkflowRunFetchedMsg struct {
	run *models.AppInstallActionWorkflowRun
	err error
}
