package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// for statuses
var (
	Pending        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	NotAttempted   = TextDim
	Approved       = TextSuccess
	ApprovalDenied = lipgloss.NewStyle().Foreground(ErrorColor)
	TerminalBad    = lipgloss.NewStyle().Foreground(WarningColor)
)

// StatusStyleMap colors AppStatus values consistently with the dashboard
// buckets in services/dashboard-ui/client/utils/status-utils.ts. Anything
// missing falls through to TextDim via GetStatusStyle.
var StatusStyleMap = map[models.AppStatus]lipgloss.Style{
	// success bucket
	models.AppStatusSuccess:                             TextSuccess,
	models.AppStatusApproved:                            Approved,
	models.AppStatusActive:                              TextSuccess,
	models.AppStatusNoDashDrift:                         TextSuccess,
	models.AppStatus(models.AppOperationStatusFinished): TextSuccess,

	// error bucket
	models.AppStatusError: TextError,

	// warn bucket
	models.AppStatusWarning:              TerminalBad,
	models.AppStatusApprovalDashAwaiting: TerminalBad,
	models.AppStatusApprovalDashDenied:   ApprovalDenied,
	models.AppStatusApprovalDashExpired:  TerminalBad,
	models.AppStatusApprovalDashRetry:    TerminalBad,
	models.AppStatusCancelled:            TerminalBad,
	models.AppStatusOutdated:             TerminalBad,
	models.AppStatusDrifted:              TerminalBad,
	models.AppStatusExpired:              TerminalBad,

	// pending / neutral bucket
	models.AppStatusPending: Pending,
	models.AppStatusNoop:    Pending,

	// in-progress bucket
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

	// skipped bucket
	models.AppStatusAutoDashSkipped: TextInfo,
	models.AppStatusUserDashSkipped: TextInfo,

	// inert / brand bucket
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

// GetRunStatusIcon returns a unicode icon for an action run status string.
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

// GetRunStatusStyle returns a lipgloss style for an action run status string.
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
