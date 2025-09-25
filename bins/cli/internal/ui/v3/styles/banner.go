package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var ApprovalConfirmation = lipgloss.NewStyle().Padding(1).
	Foreground(TextColor).
	Background(WarningColor)

var SuccessBanner = lipgloss.NewStyle().Padding(1).
	Foreground(TextColor).
	Background(SuccessColor)
