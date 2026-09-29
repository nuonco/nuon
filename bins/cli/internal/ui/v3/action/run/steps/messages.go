package steps

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type logsFetchedMsg struct {
	logs      []*models.AppOtelLogRecord
	logStream *models.AppLogStream
	err       error
}
