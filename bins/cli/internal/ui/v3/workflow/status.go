package workflow

import (
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var terminalStatuses = []models.AppStatus{
	models.AppStatusCancelled,
	models.AppStatusError,
	models.AppStatusSuccess,
	// models.AppStatusFailed
}

func getStatusIcon(status models.AppStatus) string {
	return common.GetStatusIcon(status)
}
