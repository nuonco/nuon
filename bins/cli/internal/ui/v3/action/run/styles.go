package run

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

var (
	// Border styles for different focus states
	appStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.PrimaryColor)

	appStyleBlur = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.BorderInactiveColor)

	appStyleFocus = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.BorderActiveColor)

	// Step status styles
	stepStylePending = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("8"))

	stepStyleRunning = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("11"))

	stepStyleSuccess = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("10"))

	stepStyleFailed = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("9"))

	stepStyleCancelled = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("8"))

	// Step status styles when selected
	stepStylePendingSelected = lipgloss.NewStyle().
					BorderStyle(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("8"))

	stepStyleRunningSelected = lipgloss.NewStyle().
					BorderStyle(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("11"))

	stepStyleSuccessSelected = lipgloss.NewStyle().
					BorderStyle(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("10"))

	stepStyleFailedSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("9"))

	stepStyleCancelledSelected = lipgloss.NewStyle().
					BorderStyle(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("8"))
)

// getStepStyle returns the appropriate style based on status and selection state
func getStepStyle(status string, selected bool) lipgloss.Style {
	if selected {
		switch status {
		case "running", "in_progress":
			return stepStyleRunningSelected
		case "success", "completed":
			return stepStyleSuccessSelected
		case "failed", "error":
			return stepStyleFailedSelected
		case "cancelled":
			return stepStyleCancelledSelected
		default:
			return stepStylePendingSelected
		}
	}

	switch status {
	case "running", "in_progress":
		return stepStyleRunning
	case "success", "completed":
		return stepStyleSuccess
	case "failed", "error":
		return stepStyleFailed
	case "cancelled":
		return stepStyleCancelled
	default:
		return stepStylePending
	}
}

// getStepStatusIcon returns the icon for a given status
func getStepStatusIcon(status string) string {
	switch status {
	case "running", "in_progress":
		return "⏳"
	case "success", "completed":
		return "✓"
	case "failed", "error":
		return "✗"
	case "cancelled":
		return "⊘"
	default:
		return "○"
	}
}
