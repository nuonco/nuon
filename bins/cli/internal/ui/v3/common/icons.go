package common

import "github.com/nuonco/nuon/sdks/nuon-go/models"

var StatusIconMap = map[models.AppStatus]string{
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

func GetStatusIcon(status models.AppStatus) string {
	icon, ok := StatusIconMap[status]
	if !ok {
		return "∙"
	}
	return icon
}
