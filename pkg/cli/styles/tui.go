package styles

import (
	"charm.land/lipgloss/v2"
)

// Pane border styles — the universal split-pane frame used by every alt-screen TUI.
var (
	Pane = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(PrimaryColor)

	PaneBlur = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderInactiveColor)

	PaneFocus = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderActiveColor)
)

// Step status border styles (normal and selected variants).
var (
	StepPending = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))

	StepRunning = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("11"))

	StepSuccess = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("10"))

	StepFailed = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("9"))

	StepCancelled = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))

	StepPendingSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("8"))

	StepRunningSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("11"))

	StepSuccessSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("10"))

	StepFailedSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("9"))

	StepCancelledSelected = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("8"))
)

// GetStepStyle returns the border style for a workflow/action step by status string.
func GetStepStyle(status string, selected bool) lipgloss.Style {
	if selected {
		switch status {
		case "running", "in_progress":
			return StepRunningSelected
		case "success", "completed":
			return StepSuccessSelected
		case "failed", "error":
			return StepFailedSelected
		case "cancelled":
			return StepCancelledSelected
		default:
			return StepPendingSelected
		}
	}

	switch status {
	case "running", "in_progress":
		return StepRunning
	case "success", "completed":
		return StepSuccess
	case "failed", "error":
		return StepFailed
	case "cancelled":
		return StepCancelled
	default:
		return StepPending
	}
}

// GetStepStatusIcon returns a unicode icon for a workflow/action step status.
func GetStepStatusIcon(status string) string {
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
