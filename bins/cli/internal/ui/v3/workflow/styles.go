package workflow

import (
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var appStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.PrimaryColor)

var appStyleBlur = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderInactiveColor)

var appStyleFocus = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderActiveColor)

var policySectionStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderInactiveColor)

var policyDenyStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.ErrorColor)

var policyWarnStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.WarningColor)
