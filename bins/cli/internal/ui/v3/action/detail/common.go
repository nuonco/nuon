package detail

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// TODO: consolidate in styles

func getRunStatusIcon(status string) string {
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

func getRunStatusStyle(status string) lipgloss.Style {
	switch status {
	case "success", "finished":
		return styles.TextSuccess
	case "error", "failed":
		return styles.TextError
	case "in_progress":
		return styles.TextInfo
	case "pending":
		return styles.TextDim
	case "cancelled":
		return styles.TextWarning
	default:
		return styles.TextDim
	}
}
