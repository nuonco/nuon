package logs

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var logTableStyles = table.Styles{
	Selected: lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#9d4ded")).Foreground(lipgloss.Color("#ffffff")),
	Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
	Cell:     lipgloss.NewStyle().Padding(0, 1),
}

var columns = []table.Column{
	{Title: "", Width: 0},
	{Title: "", Width: 0},
	{Title: "Level", Width: 7},
	{Title: "Timestamp", Width: 30},
	{Title: "Service", Width: 8},
	{Title: "Body", Width: 100},
}

func (m model) initTable() table.Model {
	rows := []table.Row{
		{"id", "index", "", "", "", ""},
	}
	totalWidth := 0
	for _, col := range columns {
		totalWidth += col.Width + 2
	}

	table := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(15),
		table.WithFocused(true),
		table.WithStyles(logTableStyles),
		table.WithWidth(totalWidth),
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
	m.loading = true
	logs := map[string]*models.AppOtelLogRecord{}
	if m.searchTerm != "" {
		m.setMessage(fmt.Sprintf("applying search term: %s", m.searchTerm), "info")
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
	columns := m.table.Columns()
	otherColsTotalWidth := 0
	for i, col := range columns {
		if i != len(columns)-1 {
			otherColsTotalWidth += col.Width
		}
	}
	bodyColWidth := m.table.Width() - otherColsTotalWidth - (4 * 2)

	columns[len(columns)-1].Width = bodyColWidth

	m.table.SetColumns(columns)
}
