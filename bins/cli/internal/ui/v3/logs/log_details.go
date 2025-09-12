package logs

import (
	"math"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
)

var readOnlyTableStyles = table.Styles{
	Selected: logModalBase,
	Header:   logModalBase.Bold(true).Padding(0, 1, 1, 0),
	Cell:     logModalBase.Padding(0, 2),
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
	)
	return resourceAttributesTable
}

func (m model) getLogBodyViewport(body string) viewport.Model {
	bodyViewportWidth := m.details.Width - 4
	rows := float64(len(body)) / float64(bodyViewportWidth)
	bodyViewportHeight := int(math.Ceil(rows))
	vp := viewport.New(bodyViewportWidth, bodyViewportHeight)
	vp.SetContent(body)
	return vp
}

func (m model) getDetailContent() string {
	s := ""
	s += dimTitle.Render("Body:") + "\n"
	vp := m.getLogBodyViewport(m.selectedLog.Body)
	s += logText.Render(vp.View()) + "\n\n"

	// resource attributes table
	resourceAttributesTable := m.getResourceAttributesTable()
	s += "\n" + dimTitle.Render("Resource Attributes:") + "\n"
	s += logTable.Render(resourceAttributesTable.View()) + "\n"

	// log attributes table
	logAttributesTable := m.getLogAttributesTable()
	s += "\n" + dimTitle.Render("Log Attributes:") + "\n"
	s += logTable.Render(logAttributesTable.View()) + "\n"

	return "\n" + s + "\n"
}
