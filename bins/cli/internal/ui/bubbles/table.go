package bubbles

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"
	"golang.org/x/term"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

type TableModel struct {
	table       table.Model
	quitting    bool
	interactive bool
	altScreen   bool
}

func NewTableModel(data [][]string) TableModel {
	if len(data) == 0 {
		return TableModel{}
	}

	headers := data[0]
	rows := data[1:]

	columns := make([]table.Column, len(headers))
	for i, header := range headers {
		columns[i] = table.Column{
			Title: header,
			Width: calculateColumnWidth(data, i),
		}
	}

	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRow := make(table.Row, len(headers))
		for j, cell := range row {
			if j < len(headers) {
				tableRow[j] = cell
			}
		}
		for j := len(row); j < len(headers); j++ {
			tableRow[j] = ""
		}
		tableRows[i] = tableRow
	}

	totalWidth := 0
	for _, col := range columns {
		totalWidth += col.Width + 2
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(false),
		table.WithWidth(totalWidth),
		table.WithHeight(len(tableRows)+1),
	)

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

func NewInteractiveTableModel(data [][]string) TableModel {
	model := NewTableModel(data)
	model.interactive = true
	model.altScreen = true
	return model
}

func calculateColumnWidth(data [][]string, columnIndex int) int {
	maxWidth := 10

	for _, row := range data {
		if columnIndex < len(row) {
			cellWidth := lipgloss.Width(row[columnIndex])
			if cellWidth > maxWidth {
				maxWidth = cellWidth
			}
		}
	}

	if maxWidth > 40 {
		maxWidth = 40
	}

	return maxWidth
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 4)
		if m.interactive {
			m.table.SetHeight(msg.Height - 4)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	v := tea.NewView(BaseStyle.Render(m.table.View()))
	if m.altScreen {
		v.AltScreen = true
	}
	return v
}

type TableView struct{}

func NewTableView() *TableView {
	return &TableView{}
}

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
	fmt.Println(table.viewString())
}

func (m TableModel) viewString() string {
	return BaseStyle.Render(m.table.View())
}

func (v *TableView) RenderPaging(data [][]string, offset, limit int, hasMore bool) {
	v.RenderPagingWithContext(data, offset, limit, hasMore, "", "")
}

func (v *TableView) RenderPagingWithContext(data [][]string, offset, limit int, hasMore bool, contextLabel, contextValue string) {
	v.Render(data)

	moreText := "no more items available"
	if hasMore {
		moreText = "more items available"
	}
	pagingInfo := fmt.Sprintf("offset %d, limit %d, %s", offset, limit, moreText)

	textStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true)
	footerStyle := lipgloss.NewStyle().
		Margin(1, 0, 0, 0)

	if contextValue != "" {
		contextText := contextLabel + ": " + contextValue
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			gap := w - len(pagingInfo) - len(contextText) - 2
			if gap >= 2 {
				line := textStyle.Render(pagingInfo) + strings.Repeat(" ", gap) + textStyle.Render(contextText)
				fmt.Println(footerStyle.Render(line))
				return
			}
		}
	}

	fmt.Println(footerStyle.Render(textStyle.Render(pagingInfo)))
}

func (v *TableView) RenderTotal(data [][]string, total int) {
	v.Render(data)

	textStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true)
	footerStyle := lipgloss.NewStyle().
		Margin(1, 0, 0, 0)

	fmt.Println(footerStyle.Render(textStyle.Render(fmt.Sprintf("%d total", total))))
}

func (v *TableView) RenderInteractive(data [][]string, interactive bool) error {
	if len(data) == 0 {
		noItemsStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Padding(1)
		fmt.Println(noItemsStyle.Render("No items found"))
		return nil
	}

	if !interactive {
		v.Render(data)
		return nil
	}

	model := NewInteractiveTableModel(data)
	model.table.Focus()

	program := teaprogram.NewProgram(model)
	_, err := program.Run()
	return err
}

func (v *TableView) Print(msg string) {
	fmt.Println(BaseStyle.Render(msg))
}

func (v *TableView) RenderKeyValue(pairs map[string]string) {
	if len(pairs) == 0 {
		v.Print("No data available")
		return
	}

	data := [][]string{{"Key", "Value"}}
	for key, value := range pairs {
		data = append(data, []string{key, value})
	}

	v.Render(data)
}

func RenderMarkdownTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	var result strings.Builder

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

	result.WriteString("|")
	for i, header := range headers {
		result.WriteString(fmt.Sprintf(" %-*s |", widths[i], header))
	}
	result.WriteString("\n")

	result.WriteString("|")
	for _, width := range widths {
		result.WriteString(strings.Repeat("-", width+2) + "|")
	}
	result.WriteString("\n")

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
