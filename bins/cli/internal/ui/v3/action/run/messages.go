package run

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// tickMsg is sent periodically to trigger data refresh
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second * 5)
	return tickMsg{}
}

// installActionWorkflowRunFetchedMsg is sent when the run data is fetched
type installActionWorkflowRunFetchedMsg struct {
	run *models.AppInstallActionWorkflowRun
	err error
}
