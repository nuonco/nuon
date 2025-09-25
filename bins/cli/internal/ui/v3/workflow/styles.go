package workflow

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
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
