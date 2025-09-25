package workflow

import "github.com/nuonco/nuon-go/models"

var terminalStatuses = []models.AppStatus{
	models.AppStatusCancelled,
	models.AppStatusError,
	models.AppStatusSuccess,
	// models.AppStatusFailed
}

var statusIconMap = map[models.AppStatus]string{
	models.AppStatusPending:              "⏲",
	models.AppStatusApprovalDashAwaiting: "⚠",
	models.AppStatusSuccess:              "✓",
	models.AppStatusApproved:             "✓",
	models.AppStatusCancelled:            "⊗",
	models.AppStatusError:                "⊗",
	models.AppStatusAutoDashSkipped:      "→",
	models.AppStatusUserDashSkipped:      "→",
	models.AppStatusInDashProgress:       "→",
}

func getStatusIcon(status models.AppStatus) string {
	icon, ok := statusIconMap[status]
	if !ok {
		return "∙"
	}
	return icon
}
