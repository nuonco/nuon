package steps

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// tickMsg is sent periodically to trigger data refresh
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(dataRefreshInterval)
	return tickMsg{}
}

// logsFetchedMsg is sent when logs are fetched from the API
type logsFetchedMsg struct {
	logs      []*models.AppOtelLogRecord
	logStream *models.AppLogStream
	err       error
}
