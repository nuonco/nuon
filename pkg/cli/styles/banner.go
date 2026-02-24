package styles

import (
	"charm.land/lipgloss/v2"
)

var ApprovalConfirmation = lipgloss.NewStyle().Padding(1).
	Foreground(TextColor).
	Background(WarningColor)

var SuccessBanner = lipgloss.NewStyle().Padding(1).
	Foreground(TextColor).
	Background(SuccessColor)
