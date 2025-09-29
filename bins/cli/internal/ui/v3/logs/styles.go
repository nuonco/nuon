package logs

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

var appStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.PrimaryColor)

// header holds a "title card" for the view or the search box
var headerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderInactiveColor).
	Padding(0, 1, 0, 1)

var headerStyleActive = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(styles.BorderActiveColor).
	Padding(0, 1, 0, 1)

var logTable = logModalBase.Padding(1)

var dimTitle = logModalBase.Bold(true).
	Foreground(styles.Dim)

var logText = logModalBase.
	Padding(1).
	BorderStyle(lipgloss.NormalBorder())

// Log Detail Modal
var logModalBase = lipgloss.NewStyle()
var logModal = logModalBase.
	BorderStyle(lipgloss.NormalBorder())
