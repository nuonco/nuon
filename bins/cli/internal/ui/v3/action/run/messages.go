package run

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// installActionWorkflowRunFetchedMsg is sent when the run data is fetched
type installActionWorkflowRunFetchedMsg struct {
	run *models.AppInstallActionWorkflowRun
	err error
}
