package workflow

import (
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var (
	appStyle      = styles.Pane
	appStyleBlur  = styles.PaneBlur
	appStyleFocus = styles.PaneFocus
)

// Domain-specific styles for workflow policy violations.
var policySectionStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderInactiveColor)

var policyDenyStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.ErrorColor)

var policyWarnStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.WarningColor)
