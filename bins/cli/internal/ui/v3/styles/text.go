package styles

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/nuonco/nuon-go/models"
)

var Link = lipgloss.NewStyle().Foreground(lipgloss.Color("20")).Underline(true)

var TextBold = lipgloss.NewStyle().Bold(true)
var TextDim = lipgloss.NewStyle().Foreground(lipgloss.Color("97"))
var TextLight = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
var TextSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
var TextError = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#ffffff")).
	Background(lipgloss.Color("1"))

// for statuses
var Pending = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
var NotAttempted = TextDim
var Approved = TextSuccess
var ApprovalDenied = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
var Cancelled = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
var Error = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
var Info = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))

var StatusStyleMap = map[models.AppStatus]lipgloss.Style{
	models.AppStatusPending:              Pending,
	models.AppStatusNotDashAttempted:     NotAttempted,
	models.AppStatusApproved:             Approved,
	models.AppStatusApprovalDashDenied:   ApprovalDenied,
	models.AppStatusCancelled:            Cancelled,
	models.AppStatusError:                Error,
	models.AppStatusAutoDashSkipped:      Info,
	models.AppStatusApprovalDashAwaiting: Cancelled,
	models.AppStatusSuccess:              TextSuccess,
}

func GetStatusStyle(status models.AppStatus) lipgloss.Style {
	style, ok := StatusStyleMap[status]
	if ok {
		return style
	}
	return TextDim
}

var ApprovalConfirmation = lipgloss.NewStyle().Padding(1).
	Foreground(lipgloss.Color("17")).
	Background(lipgloss.Color("11"))

var SuccessBanner = lipgloss.NewStyle().Padding(1).
	Foreground(lipgloss.Color("17")).
	Background(lipgloss.Color("36"))
