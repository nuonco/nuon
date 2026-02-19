package styles

import (
	"github.com/charmbracelet/lipgloss"

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
