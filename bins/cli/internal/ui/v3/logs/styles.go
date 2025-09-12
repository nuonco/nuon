package logs

import "github.com/charmbracelet/lipgloss"

var appStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#6133FF"))

// header holds a "title card" for the view or the search box
var headerStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1, 0, 1)

var headerStyleActive = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#6133FF")).
	Padding(0, 1, 0, 1)

// Log Detail Modal
var logModalBase = lipgloss.NewStyle().Background(lipgloss.Color("#260c4f"))
var logModalHeader = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder())
var logModal = logModalBase.
	BorderStyle(lipgloss.RoundedBorder()).
	Padding(1).
	Align(lipgloss.Center, lipgloss.Center)
var dimTitle = logModalBase.Bold(true).
	Foreground(lipgloss.Color("#F9519D"))
var logText = logModalBase.
	Padding(1).
	BorderStyle(lipgloss.NormalBorder()).
	Background(lipgloss.Color("#000000"))
var logTable = logModalBase.Padding(1)

// message footer
var messageStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("240"))
