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

var StatusStyleMap = map[models.AppStatus]lipgloss.Style{
	models.AppStatusPending:                             Pending,
	models.AppStatusNotDashAttempted:                    TextDefault,
	models.AppStatusApproved:                            Approved,
	models.AppStatusApprovalDashDenied:                  ApprovalDenied,
	models.AppStatusCancelled:                           TerminalBad,
	models.AppStatusError:                               lipgloss.NewStyle().Foreground(ErrorColor),
	models.AppStatusAutoDashSkipped:                     lipgloss.NewStyle().Foreground(InfoColor),
	models.AppStatusApprovalDashAwaiting:                TerminalBad,
	models.AppStatusSuccess:                             TextSuccess,
	models.AppStatus(models.AppOperationStatusFinished): TextSuccess,
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
