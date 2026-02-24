package steps

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// logsFetchedMsg is sent when logs are fetched from the API
type logsFetchedMsg struct {
	logs      []*models.AppOtelLogRecord
	logStream *models.AppLogStream
	err       error
}
