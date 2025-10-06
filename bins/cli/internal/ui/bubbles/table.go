package bubbles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// TableModel represents a data table component
type TableModel struct {
	table       table.Model
	quitting    bool
	interactive bool
}

// NewTableModel creates a new table model
func NewTableModel(data [][]string) TableModel {
	if len(data) == 0 {
		return TableModel{}
	}

	// First row is headers
	headers := data[0]
	rows := data[1:]

	// Create columns from headers
	columns := make([]table.Column, len(headers))
	for i, header := range headers {
		columns[i] = table.Column{
			Title: header,
			Width: calculateColumnWidth(data, i),
		}
	}

	// Create table rows
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		// Ensure row has same length as headers
		tableRow := make(table.Row, len(headers))
		for j, cell := range row {
			if j < len(headers) {
				tableRow[j] = cell
			}
		}
		// Fill empty cells if row is shorter than headers
		for j := len(row); j < len(headers); j++ {
			tableRow[j] = ""
		}
		tableRows[i] = tableRow
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(false),
		table.WithHeight(len(tableRows)+1), // Set height to match number of rows + header
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.PrimaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(styles.PrimaryColor)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("")).
		Background(lipgloss.Color("")).
		Bold(false)

	t.SetStyles(s)

	return TableModel{
		table:       t,
		interactive: false,
	}
}

// NewInteractiveTableModel creates a new interactive table model
func NewInteractiveTableModel(data [][]string) TableModel {
	model := NewTableModel(data)
	model.interactive = true
	return model
}

// calculateColumnWidth determines appropriate column width
func calculateColumnWidth(data [][]string, columnIndex int) int {
	maxWidth := 10 // minimum width

	for _, row := range data {
		if columnIndex < len(row) {
			cellWidth := len(row[columnIndex])
			if cellWidth > maxWidth {
				maxWidth = cellWidth
			}
		}
	}

	// Cap maximum width
	if maxWidth > 40 {
		maxWidth = 40
	}

	return maxWidth
}

// Init initializes the table model
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the table model
func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 4) // Account for padding
		// Only set height to full screen for interactive mode
		if m.interactive {
			m.table.SetHeight(msg.Height - 4)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table
func (m TableModel) View() string {
	if m.quitting {
		return ""
	}

	return BaseStyle.Render(m.table.View())
}

// TableView provides a high-level interface for rendering tables
type TableView struct{}

// NewTableView creates a new table view
func NewTableView() *TableView {
	return &TableView{}
}

// Render displays a table with the given data
func (v *TableView) Render(data [][]string) {
	if len(data) == 0 {
		noItemsStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Padding(1)
		fmt.Println(noItemsStyle.Render("No items found"))
		return
	}

	table := NewTableModel(data)
	fmt.Println(table.View())
}

// RenderPaging displays a table with pagination information
func (v *TableView) RenderPaging(data [][]string, offset, limit int, hasMore bool) {
	v.Render(data)

	// Add pagination info
	pagingStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true).
		Margin(1, 0, 0, 0)

	moreText := "no more items available"
	if hasMore {
		moreText = "more items available"
	}

	pagingInfo := fmt.Sprintf("offset %d, limit %d, %s", offset, limit, moreText)
	fmt.Println(pagingStyle.Render(pagingInfo))
}

// RenderInteractive displays an interactive table that users can navigate
func (v *TableView) RenderInteractive(data [][]string) error {
	if len(data) == 0 {
		noItemsStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Padding(1)
		fmt.Println(noItemsStyle.Render("No items found"))
		return nil
	}

	model := NewInteractiveTableModel(data)
	model.table.Focus()

	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// Print displays a simple message
func (v *TableView) Print(msg string) {
	fmt.Println(BaseStyle.Render(msg))
}

// RenderKeyValue renders key-value pairs in a table format
func (v *TableView) RenderKeyValue(pairs map[string]string) {
	if len(pairs) == 0 {
		v.Print("No data available")
		return
	}

	// Convert to table data
	data := [][]string{{"Key", "Value"}}
	for key, value := range pairs {
		data = append(data, []string{key, value})
	}

	v.Render(data)
}

// RenderMarkdown renders a simple markdown-like table
func RenderMarkdownTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	var result strings.Builder

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Header
	result.WriteString("|")
	for i, header := range headers {
		result.WriteString(fmt.Sprintf(" %-*s |", widths[i], header))
	}
	result.WriteString("\n")

	// Separator
	result.WriteString("|")
	for _, width := range widths {
		result.WriteString(strings.Repeat("-", width+2) + "|")
	}
	result.WriteString("\n")

	// Data rows
	for _, row := range rows {
		result.WriteString("|")
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			result.WriteString(fmt.Sprintf(" %-*s |", widths[i], cell))
		}
		result.WriteString("\n")
	}

	return result.String()
}
