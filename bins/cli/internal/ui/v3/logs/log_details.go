package logs

import (
	"math"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	// "github.com/powertoolsdev/mono/pkg/cli/styles"
)

var readOnlyTableStyles = table.Styles{
	Selected: logModalBase,
	Header:   logModalBase.Bold(true).Padding(0, 1, 1, 0),
	Cell:     logModalBase.Padding(0, 1),
}

func (m model) getLogAttributesTable() table.Model {
	lckw := 0
	lcvw := 0
	logColumns := []table.Column{
		{Title: "Key", Width: 10},
		{Title: "Value", Width: 50},
	}
	logAttributesRows := []table.Row{}
	for k, v := range m.selectedLog.LogAttributes {
		if len(k) > lckw {
			lckw = len(k)
		}
		if len(v) > lcvw {
			lcvw = len(v)
		}
		logAttributesRows = append(logAttributesRows, table.Row{k, v})
	}
	logColumns[0].Width = lckw
	logColumns[1].Width = lcvw

	logAttributesTable := table.New(
		table.WithColumns(logColumns),
		table.WithRows(logAttributesRows),
		table.WithHeight(len(logAttributesRows)),
		table.WithFocused(true),
		table.WithStyles(readOnlyTableStyles),
		table.WithFocused(false),
		table.WithWidth(m.sidebarWidth),
	)
	return logAttributesTable
}

func (m model) getResourceAttributesTable() table.Model {
	rckw := 0
	rcvw := 0
	resourceColumns := []table.Column{
		{Title: "Key", Width: 10},
		{Title: "Value", Width: 50},
	}
	resourceAttributesRows := []table.Row{}
	for k, v := range m.selectedLog.ResourceAttributes {
		if len(k) > rckw {
			rckw = len(k)
		}
		if len(v) > rcvw {
			rcvw = len(v)
		}

		resourceAttributesRows = append(resourceAttributesRows, table.Row{k, v})
	}
	resourceColumns[0].Width = rckw
	resourceColumns[1].Width = rcvw

	resourceAttributesTable := table.New(
		table.WithColumns(resourceColumns),
		table.WithRows(resourceAttributesRows),
		table.WithHeight(len(resourceAttributesRows)),
		table.WithFocused(true),
		table.WithStyles(readOnlyTableStyles),
		table.WithFocused(false),
		table.WithWidth(m.sidebarWidth),
	)
	return resourceAttributesTable
}

func (m model) getLogBody(body string) string {
	// returns a body within a box of the right size.
	width := m.sidebarWidth - 4 // minus whitespace
	height := int(math.Ceil(float64(len(body)) / float64(width)))
	return logText.Width(width).Height(height).Render(body)
}

func (m model) getDetailContent() string {
	sections := []string{}
	// body
	sections = append(sections,
		lipgloss.NewStyle().Width(m.sidebarWidth).Padding(1).Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				dimTitle.Width(m.sidebarWidth-2).Render("Body:"),
				m.getLogBody(m.selectedLog.Body),
			),
		),
	)
	// resource attributes table
	sections = append(sections,
		lipgloss.NewStyle().Width(m.sidebarWidth-2).Padding(1).Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				dimTitle.Render("Resource Attributes:"),
				logTable.Render(m.getResourceAttributesTable().View()),
			),
		),
	)

	// log attributes table
	sections = append(sections,
		lipgloss.NewStyle().Width(m.sidebarWidth-2).Padding(1).Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				dimTitle.Render("Log Attributes:"),
				logTable.Render(m.getLogAttributesTable().View()),
			),
		),
	)

	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
