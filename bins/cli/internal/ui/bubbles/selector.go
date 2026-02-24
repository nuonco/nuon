package bubbles

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// SelectorItem represents an item in the selector list
type SelectorItem struct {
	title        string
	description  string
	value        string
	isEvaluation bool // Special marking for evaluation journey items
}

// Implement list.Item interface
func (i SelectorItem) FilterValue() string { return i.title }
func (i SelectorItem) Title() string       { return i.title }
func (i SelectorItem) Description() string { return i.description }
func (i SelectorItem) Value() string       { return i.value }
func (i SelectorItem) IsEvaluation() bool  { return i.isEvaluation }

type searchDebounceMsg struct{ query string }
type searchResultMsg struct {
	items []SelectorItem
	err   error
}

// SelectorModel represents the list selection component
type SelectorModel struct {
	items          []SelectorItem
	filteredItems  []SelectorItem
	originalItems  []SelectorItem
	searchFn       func(string) ([]SelectorItem, error)
	searching      bool
	searchErr      string
	choice         string
	selected       bool
	quitting       bool
	cursor         int
	width          int
	height         int
	searchQuery    string
	searchMode     bool
	maxVisibleRows int // Maximum number of rows to display at once, 0 = auto-calculate
	viewportOffset int // Scroll offset for visible items
}

// NewSelectorModel creates a new selector model
func NewSelectorModel(title string, items []SelectorItem) SelectorModel {
	return SelectorModel{
		items:          items,
		filteredItems:  items,
		originalItems:  items,
		cursor:         0,
		width:          60,
		height:         24, // Default terminal height
		searchQuery:    "",
		searchMode:     false,
		maxVisibleRows: 0, // Auto-calculate based on terminal height
		viewportOffset: 0,
	}
}

// NewSelectorModelWithMaxRows creates a new selector model with a specific max visible rows
func NewSelectorModelWithMaxRows(title string, items []SelectorItem, maxVisibleRows int) SelectorModel {
	model := NewSelectorModel(title, items)
	model.maxVisibleRows = maxVisibleRows
	return model
}

// Init initializes the selector model
func (m SelectorModel) Init() tea.Cmd {
	return nil
}

// filterItems filters items based on the search query using fuzzy matching
func (m *SelectorModel) filterItems() {
	if m.searchQuery == "" {
		m.filteredItems = m.items
		return
	}

	var filtered []SelectorItem
	for _, item := range m.items {
		searchText := item.Title()
		if item.Description() != "" {
			searchText += " " + item.Description()
		}
		if fuzzy.MatchFold(m.searchQuery, searchText) {
			filtered = append(filtered, item)
		}
	}

	m.filteredItems = filtered
	if m.cursor >= len(m.filteredItems) {
		m.cursor = len(m.filteredItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.adjustViewport()
}

// adjustViewport ensures the cursor is visible within the viewport
func (m *SelectorModel) adjustViewport() {
	visibleRows := m.getVisibleRows()
	if visibleRows <= 0 || len(m.filteredItems) <= visibleRows {
		m.viewportOffset = 0
		return
	}
	if m.cursor >= m.viewportOffset+visibleRows {
		m.viewportOffset = m.cursor - visibleRows + 1
	}
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}
	maxOffset := len(m.filteredItems) - visibleRows
	if m.viewportOffset > maxOffset {
		m.viewportOffset = max(0, maxOffset)
	}
}

// getVisibleRows calculates the number of rows that can be displayed
func (m *SelectorModel) getVisibleRows() int {
	if m.maxVisibleRows > 0 {
		return m.maxVisibleRows
	}
	reservedLines := 9
	availableHeight := m.height - reservedLines
	if availableHeight < 5 {
		availableHeight = 5
	}
	return availableHeight
}

func searchDebounceCmd(query string) tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return searchDebounceMsg{query: query}
	})
}

// Update handles messages for the selector model
func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 64 {
			m.width = 60
		} else {
			m.width = msg.Width - 4
		}
		m.height = msg.Height
		m.adjustViewport()
		return m, nil

	case searchDebounceMsg:
		if msg.query != m.searchQuery || m.searchFn == nil {
			return m, nil
		}
		m.searching = true
		m.searchErr = ""
		fn := m.searchFn
		q := msg.query
		return m, func() tea.Msg {
			items, err := fn(q)
			return searchResultMsg{items: items, err: err}
		}

	case searchResultMsg:
		m.searching = false
		if msg.err != nil {
			m.searchErr = "Search failed"
			return m, nil
		}
		m.filteredItems = msg.items
		m.cursor = 0
		m.viewportOffset = 0
		return m, nil

	case tea.KeyPressMsg:
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				return m, nil
			case "enter":
				m.searchMode = false
				if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
					m.choice = m.filteredItems[m.cursor].Value()
					m.selected = true
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					if m.searchFn != nil {
						if m.searchQuery == "" {
							m.searching = false
							m.searchErr = ""
							m.filteredItems = m.originalItems
							m.cursor, m.viewportOffset = 0, 0
							return m, nil
						}
						return m, searchDebounceCmd(m.searchQuery)
					}
					m.filterItems()
				}
				return m, nil
			case "up":
				if m.cursor > 0 {
					m.cursor--
					m.adjustViewport()
				}
				return m, nil
			case "down":
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
					m.adjustViewport()
				}
				return m, nil
			case "space":
				m.searchQuery += " "
				if m.searchFn != nil {
					return m, searchDebounceCmd(m.searchQuery)
				}
				m.filterItems()
				return m, nil
			default:
				if text := msg.Key().Text; text != "" {
					m.searchQuery += text
					if m.searchFn != nil {
						return m, searchDebounceCmd(m.searchQuery)
					}
					m.filterItems()
				}
				return m, nil
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.quitting = true
				return m, tea.Quit
			case "up":
				if m.cursor > 0 {
					m.cursor--
					m.adjustViewport()
				}
				return m, nil
			case "down":
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
					m.adjustViewport()
				}
				return m, nil
			case "enter":
				if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
					m.choice = m.filteredItems[m.cursor].Value()
					m.selected = true
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil
			default:
				if text := msg.Key().Text; text != "" {
					if text == "/" {
						m.searchMode = true
						return m, nil
					}
					m.searchMode = true
					m.searchQuery = text
					if m.searchFn != nil {
						return m, searchDebounceCmd(m.searchQuery)
					}
					m.filterItems()
					return m, nil
				}
			}
		}
	}

	return m, nil
}

// View renders the selector
func (m SelectorModel) View() tea.View {
	if m.quitting {
		if m.selected {
			if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
				selectedItem := m.filteredItems[m.cursor]
				successStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
				return tea.NewView(successStyle.Render(fmt.Sprintf("✓ Selected: %s", selectedItem.Title())))
			}
		}
		return tea.NewView("")
	}

	var b strings.Builder

	searchBoxStyle := lipgloss.NewStyle().
		Foreground(styles.TextColor).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Width(m.width - 2)

	searchPrompt := ">"
	if m.searchMode {
		searchBoxStyle = searchBoxStyle.BorderForeground(styles.PrimaryColor)
	}

	searchText := m.searchQuery
	if searchText == "" && !m.searchMode {
		searchText = "Type press / to search..."
		searchBoxStyle = searchBoxStyle.Foreground(styles.SubtleColor)
	}

	if m.searching {
		searchText = "Searching…"
		searchBoxStyle = searchBoxStyle.Foreground(styles.SubtleColor)
	} else if m.searchMode {
		searchText = m.searchQuery + "█"
	}

	b.WriteString(searchBoxStyle.Render(fmt.Sprintf("%s %s", searchPrompt, searchText)))
	b.WriteString("\n")

	if len(m.filteredItems) == 0 {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Align(lipgloss.Center).
			Padding(2, 0)
		msg := "No matches found"
		if m.searchErr != "" {
			msg = m.searchErr
		} else if m.searching {
			msg = ""
		}
		b.WriteString(noResultsStyle.Render(msg))
		b.WriteString("\n")
	} else {
		visibleRows := m.getVisibleRows()
		startIdx := m.viewportOffset
		endIdx := min(startIdx+visibleRows, len(m.filteredItems))

		if startIdx > 0 {
			scrollIndicatorStyle := lipgloss.NewStyle().
				Foreground(styles.SubtleColor).
				Italic(true)
			b.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more above...", startIdx)))
			b.WriteString("\n")
		}

		for i := startIdx; i < endIdx; i++ {
			item := m.filteredItems[i]
			var itemStyle lipgloss.Style
			prefix := "  "

			if i == m.cursor {
				itemStyle = lipgloss.NewStyle().
					Foreground(styles.PrimaryColor).
					Bold(true)
				prefix = "▶ "
			} else {
				itemStyle = lipgloss.NewStyle().
					Foreground(styles.TextColor)
			}

			line := fmt.Sprintf("%s%s", prefix, item.Title())
			if item.Description() != "" {
				line = fmt.Sprintf("%s%s %s", prefix, item.Title(), item.Description())
			}

			b.WriteString(itemStyle.Render(line))
			b.WriteString("\n")
		}

		if endIdx < len(m.filteredItems) {
			scrollIndicatorStyle := lipgloss.NewStyle().
				Foreground(styles.SubtleColor).
				Italic(true)
			remainingItems := len(m.filteredItems) - endIdx
			b.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more below...", remainingItems)))
			b.WriteString("\n")
		}
	}

	if m.searchQuery != "" {
		countStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Margin(1, 0, 0, 0)
		b.WriteString(countStyle.Render(fmt.Sprintf("Found %02d match(es)", len(m.filteredItems))))
		b.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true).
		Margin(1, 0, 0, 0)

	helpText := "Use ↑/↓ to navigate, Enter to select, Esc to cancel, / to search"
	if m.searchMode {
		helpText = "Type to filter, ↑/↓ to navigate, Enter to select, Esc to exit search"
	}
	b.WriteString(helpStyle.Render(helpText))

	return tea.NewView(BorderStyle.Render(b.String()))
}

// Choice returns the selected choice value
func (m SelectorModel) Choice() string { return m.choice }

// Selected returns whether a choice was made
func (m SelectorModel) Selected() bool { return m.selected }

// High-level selector functions

func SelectFromOptions(title string, options []string, interactive bool) (string, error) {
	items := make([]SelectorItem, len(options))
	for i, option := range options {
		items[i] = SelectorItem{title: option, value: option}
	}
	return SelectFromItems(title, items, interactive)
}

func SelectFromOptionsWithMaxRows(title string, options []string, maxVisibleRows int, interactive bool) (string, error) {
	items := make([]SelectorItem, len(options))
	for i, option := range options {
		items[i] = SelectorItem{title: option, value: option}
	}
	return SelectFromItemsWithMaxRows(title, items, maxVisibleRows, interactive)
}

func SelectFromItems(title string, items []SelectorItem, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use the appropriate --id flag to specify directly")
	}
	model := NewSelectorModel(title, items)
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}
	selectorModel := finalModel.(SelectorModel)
	if !selectorModel.Selected() {
		return "", fmt.Errorf("selection cancelled")
	}
	return selectorModel.Choice(), nil
}

func SelectFromItemsWithMaxRows(title string, items []SelectorItem, maxVisibleRows int, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use the appropriate --id flag to specify directly")
	}
	model := NewSelectorModelWithMaxRows(title, items, maxVisibleRows)
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}
	selectorModel := finalModel.(SelectorModel)
	if !selectorModel.Selected() {
		return "", fmt.Errorf("selection cancelled")
	}
	return selectorModel.Choice(), nil
}

func SelectOrg(orgs []OrgOption, searchFn func(string) ([]OrgOption, error), interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use --org-id flag to specify directly")
	}

	buildItems := func(opts []OrgOption) []SelectorItem {
		maxWidth := 0
		for _, org := range opts {
			if len(org.Name) > maxWidth {
				maxWidth = len(org.Name)
			}
		}
		items := make([]SelectorItem, len(opts))
		for i, org := range opts {
			title := fmt.Sprintf("%s%s", org.Name, strings.Repeat(" ", maxWidth-len(org.Name)))
			description := styles.TextDim.Render(fmt.Sprintf("ID: %s", org.ID))
			if org.IsEvaluation {
				title = fmt.Sprintf("🚀 %s (Evaluation)", org.Name)
				description = fmt.Sprintf("ID: %s • Perfect for trying out Nuon", org.ID)
			}
			items[i] = SelectorItem{title: title, description: description, value: org.ID, isEvaluation: org.IsEvaluation}
		}
		return items
	}

	items := buildItems(orgs)
	title := "Select an organization"
	if hasEvaluationOrgs(orgs) {
		title = "Select an organization (🚀 = Evaluation mode)"
	}

	var selectorSearchFn func(string) ([]SelectorItem, error)
	if searchFn != nil {
		selectorSearchFn = func(q string) ([]SelectorItem, error) {
			orgResults, err := searchFn(q)
			if err != nil {
				return nil, err
			}
			return buildItems(orgResults), nil
		}
	}

	model := NewSelectorModel(title, items)
	model.originalItems = items
	model.searchFn = selectorSearchFn
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}
	selectorModel := finalModel.(SelectorModel)
	if !selectorModel.Selected() {
		return "", fmt.Errorf("selection cancelled")
	}
	return selectorModel.Choice(), nil
}

func SelectApp(apps []AppOption, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use --app-id flag to specify directly")
	}
	items := make([]SelectorItem, len(apps))
	maxAppNameWidth := 0
	for _, app := range apps {
		if len(app.Name) > maxAppNameWidth {
			maxAppNameWidth = len(app.Name)
		}
	}
	for i, app := range apps {
		items[i] = SelectorItem{
			title:       fmt.Sprintf("%s%s", app.Name, strings.Repeat(" ", maxAppNameWidth-len(app.Name))),
			description: fmt.Sprintf("ID: %s", app.ID),
			value:       app.ID,
		}
	}
	return SelectFromItems("Select an application", items, interactive)
}

func SelectInstall(installs []InstallOption, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use --install-id flag to specify directly")
	}
	items := make([]SelectorItem, len(installs))
	maxInstallNameWidth := 0
	for _, install := range installs {
		if len(install.Name) > maxInstallNameWidth {
			maxInstallNameWidth = len(install.Name)
		}
	}
	for i, install := range installs {
		items[i] = SelectorItem{
			title:       fmt.Sprintf("%s%s", install.Name, strings.Repeat(" ", maxInstallNameWidth-len(install.Name))),
			description: styles.TextDim.Render(fmt.Sprintf("ID: %s", install.ID)),
			value:       install.ID,
		}
	}
	return SelectFromItems("Select an installation", items, interactive)
}

func SelectWorkflow(workflows []WorkflowOption, interactive bool) (string, error) {
	if !interactive {
		return "", fmt.Errorf("interactive terminal required for selection; use --workflow-id flag to specify directly")
	}
	items := make([]SelectorItem, len(workflows))
	maxWorkflowNameWidth := 0
	for _, workflow := range workflows {
		if len(workflow.Name) > maxWorkflowNameWidth {
			maxWorkflowNameWidth = len(workflow.Name)
		}
	}
	for i, workflow := range workflows {
		desc := fmt.Sprintf("ID: %s", workflow.ID)
		if workflow.Type != "" {
			desc += fmt.Sprintf(" • %s", workflow.Type)
		}
		if workflow.Status != "" {
			desc += fmt.Sprintf(" • %s", workflow.Status)
		}
		items[i] = SelectorItem{
			title:       fmt.Sprintf("%s%s", workflow.Name, strings.Repeat(" ", maxWorkflowNameWidth-len(workflow.Name))),
			description: styles.TextDim.Render(desc),
			value:       workflow.ID,
		}
	}
	return SelectFromItems("Select a workflow", items, interactive)
}

type OrgOption struct {
	ID           string
	Name         string
	IsEvaluation bool
}

type AppOption struct {
	ID   string
	Name string
}

type InstallOption struct {
	ID   string
	Name string
}

type WorkflowOption struct {
	ID     string
	Name   string
	Type   string
	Status string
}

func hasEvaluationOrgs(orgs []OrgOption) bool {
	for _, org := range orgs {
		if org.IsEvaluation {
			return true
		}
	}
	return false
}

func ParseOrgSelection(selection string) (name, id string) {
	parts := strings.Split(selection, ":")
	if len(parts) >= 2 {
		name = strings.TrimSpace(parts[0])
		id = strings.TrimSpace(parts[1])
	}
	return
}
