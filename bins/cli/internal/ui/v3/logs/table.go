package logs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
)

// extends the base styles
var logTableStyles = table.Styles{
	Selected: lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#9d4ded")).Foreground(lipgloss.Color("#ffffff")),
	Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
	Cell:     lipgloss.NewStyle().Padding(0, 1),
}

var columns = []table.Column{
	{Title: "", Width: 0}, // hidden ID
	{Title: "", Width: 0}, // hidden index
	{Title: "Level", Width: 7},
	{Title: "Timestamp", Width: 30},
	{Title: "Service", Width: 8},
	{Title: "Body", Width: 100}, // NOTE(fd): this is dynamically configured during resizes
}

func (m model) initTable() table.Model {
	rows := []table.Row{
		{"id", "index", "", "", "", ""},
	}
	table := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(15),
		table.WithFocused(true),
		table.WithStyles(logTableStyles),
		// table.WithWidth(30),
	)
	return table
}

func rowFromLog(i int, log *models.AppOtelLogRecord) []string {
	// TODO(fd): style in here
	return []string{
		log.ID,
		fmt.Sprintf("%d", i),
		log.SeverityText,
		log.Timestamp,
		log.ServiceName,
		log.Body,
	}
}

func (m *model) prepareRows() []table.Row {
	// this method is hella overloaded, break it up
	m.loading = true
	logs := map[string]*models.AppOtelLogRecord{}
	// NOTE(fd): this is a naive approach
	if m.searchTerm != "" {
		m.message = fmt.Sprintf("applying search term: %s", m.searchTerm)
		filteredLogs := map[string]*models.AppOtelLogRecord{}
		for _, log := range m.logs {
			matches := strings.Contains(strings.ToLower(log.Body), strings.ToLower(m.searchTerm))
			if matches {
				filteredLogs[log.ID] = log
			}
		}
		logs = filteredLogs

	} else {
		logs = m.logs
	}

	listSize := len(logs)
	var rows = make([]table.Row, listSize)
	i := 0
	for id := range logs {
		log := logs[id]
		rows[i] = rowFromLog(i, log)
		i++
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][3] < rows[j][3]
	})

	m.loading = false
	return rows
}

func (m *model) resizeTableColumns() {
	// the body columne lengths are dynamic but we want to trim the text in case
	// the content is too long. so we must resize the body column every time
	// the viewport resizes. we do this by modifying the column in the Column list var
	// and setting the modified columns on the table.

	// NOTE(fd): it may be a good idea to make a copy of the columns instead

	// calculate the column width
	columns := m.table.Columns()
	otherColsTotalWidth := 0
	for i, col := range columns {
		if i != len(columns)-1 {
			otherColsTotalWidth += col.Width
		}
	}
	// 4 * 2 is (num cols) * padding
	bodyColWidth := m.table.Width() - otherColsTotalWidth - (4 * 2)

	// set the column width on the body column by index
	columns[len(columns)-1].Width = bodyColWidth

	// set the new columns on the table
	m.table.SetColumns(columns)
}
