package common

import "github.com/nuonco/nuon/sdks/nuon-go/models"

var StatusIconMap = map[models.AppStatus]string{
	models.AppStatusSuccess:                             "✓",
	models.AppStatusApproved:                            "✓",
	models.AppStatusActive:                              "✓",
	models.AppStatusNoDashDrift:                         "✓",
	models.AppStatus(models.AppOperationStatusFinished): "✓",

	models.AppStatusError: "⊗",

	models.AppStatusWarning:              "⚠",
	models.AppStatusApprovalDashAwaiting: "⚠",
	models.AppStatusApprovalDashDenied:   "⚠",
	models.AppStatusApprovalDashExpired:  "⚠",
	models.AppStatusApprovalDashRetry:    "⚠",
	models.AppStatusOutdated:             "⚠",
	models.AppStatusDrifted:              "⚠",
	models.AppStatusExpired:              "⚠",

	models.AppStatusCancelled: "⊗",

	models.AppStatusPending: "⏲",
	models.AppStatusNoop:    "⏲",

	models.AppStatusInDashProgress:          "→",
	models.AppStatusPlanning:                "→",
	models.AppStatusApplying:                "→",
	models.AppStatusProvisioning:            "→",
	models.AppStatusBuilding:                "→",
	models.AppStatusQueued:                  "→",
	models.AppStatusGenerating:              "→",
	models.AppStatusRetrying:                "→",
	models.AppStatusCheckingDashPlan:        "→",
	models.AppStatusAwaitingDashUserDashRun: "→",
	models.AppStatusDeleting:                "→",

	models.AppStatusAutoDashSkipped: "→",
	models.AppStatusUserDashSkipped: "→",
}

var InProgressStatuses = map[models.AppStatus]struct{}{
	models.AppStatusInDashProgress:          {},
	models.AppStatusPlanning:                {},
	models.AppStatusApplying:                {},
	models.AppStatusProvisioning:            {},
	models.AppStatusBuilding:                {},
	models.AppStatusQueued:                  {},
	models.AppStatusGenerating:              {},
	models.AppStatusRetrying:                {},
	models.AppStatusCheckingDashPlan:        {},
	models.AppStatusAwaitingDashUserDashRun: {},
	models.AppStatusDeleting:                {},
}

func IsInProgressStatus(status models.AppStatus) bool {
	_, ok := InProgressStatuses[status]
	return ok
}

func GetStatusIcon(status models.AppStatus) string {
	icon, ok := StatusIconMap[status]
	if !ok {
		return "∙"
	}
	return icon
}
