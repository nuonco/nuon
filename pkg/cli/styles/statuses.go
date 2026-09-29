package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var (
	Pending        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	NotAttempted   = TextDim
	Approved       = TextSuccess
	ApprovalDenied = lipgloss.NewStyle().Foreground(ErrorColor)
	TerminalBad    = lipgloss.NewStyle().Foreground(WarningColor)
)

var StatusStyleMap = map[models.AppStatus]lipgloss.Style{
	models.AppStatusSuccess:                             TextSuccess,
	models.AppStatusApproved:                            Approved,
	models.AppStatusActive:                              TextSuccess,
	models.AppStatusNoDashDrift:                         TextSuccess,
	models.AppStatus(models.AppOperationStatusFinished): TextSuccess,

	models.AppStatusError: TextError,

	models.AppStatusWarning:              TerminalBad,
	models.AppStatusApprovalDashAwaiting: TerminalBad,
	models.AppStatusApprovalDashDenied:   ApprovalDenied,
	models.AppStatusApprovalDashExpired:  TerminalBad,
	models.AppStatusApprovalDashRetry:    TerminalBad,
	models.AppStatusCancelled:            TerminalBad,
	models.AppStatusOutdated:             TerminalBad,
	models.AppStatusDrifted:              TerminalBad,
	models.AppStatusExpired:              TerminalBad,

	models.AppStatusPending: Pending,
	models.AppStatusNoop:    Pending,

	models.AppStatusInDashProgress:          TextInfo,
	models.AppStatusPlanning:                TextInfo,
	models.AppStatusApplying:                TextInfo,
	models.AppStatusProvisioning:            TextInfo,
	models.AppStatusBuilding:                TextInfo,
	models.AppStatusQueued:                  TextInfo,
	models.AppStatusGenerating:              TextInfo,
	models.AppStatusRetrying:                TextInfo,
	models.AppStatusCheckingDashPlan:        TextInfo,
	models.AppStatusAwaitingDashUserDashRun: TextInfo,
	models.AppStatusDeleting:                TextInfo,

	models.AppStatusAutoDashSkipped: TextInfo,
	models.AppStatusUserDashSkipped: TextInfo,

	models.AppStatusNotDashAttempted: TextDefault,
	models.AppStatusDiscarded:        TextDim,
}

func GetStatusStyle(status models.AppStatus) lipgloss.Style {
	style, ok := StatusStyleMap[status]
	if ok {
		return style
	}
	return TextDim
}

func GetRunStatusIcon(status string) string {
	switch status {
	case "success", "finished":
		return "✓"
	case "error", "failed", "cancelled":
		return "✗"
	case "in_progress", "pending":
		return "⟳"
	default:
		return "○"
	}
}

func GetRunStatusStyle(status string) lipgloss.Style {
	switch status {
	case "success", "finished":
		return TextSuccess
	case "error", "failed":
		return TextError
	case "in_progress":
		return TextInfo
	case "pending":
		return TextDim
	case "cancelled":
		return TextWarning
	default:
		return TextDim
	}
}

func IsActionableStatus(status models.AppStatus) bool {
	switch status {
	case models.AppStatusApprovalDashAwaiting,
		models.AppStatusPending,
		models.AppStatusInDashProgress,
		models.AppStatusPlanning,
		models.AppStatusApplying,
		models.AppStatusQueued,
		models.AppStatusGenerating,
		models.AppStatusRetrying,
		models.AppStatusBuilding,
		models.AppStatusProvisioning,
		models.AppStatusCheckingDashPlan:
		return true
	default:
		return false
	}
}
