package bubbles

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
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

// SelectorModel represents the list selection component
type SelectorModel struct {
	items          []SelectorItem
	filteredItems  []SelectorItem
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
		// Create searchable text from title and description
		searchText := item.Title()
		if item.Description() != "" {
			searchText += " " + item.Description()
		}

		if fuzzy.MatchFold(m.searchQuery, searchText) {
			filtered = append(filtered, item)
		}
	}

	m.filteredItems = filtered

	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filteredItems) {
		m.cursor = len(m.filteredItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Adjust viewport after filtering
	m.adjustViewport()
}

// adjustViewport ensures the cursor is visible within the viewport
func (m *SelectorModel) adjustViewport() {
	visibleRows := m.getVisibleRows()

	if visibleRows <= 0 || len(m.filteredItems) <= visibleRows {
		// No scrolling needed
		m.viewportOffset = 0
		return
	}

	// Scroll down if cursor is below viewport
	if m.cursor >= m.viewportOffset+visibleRows {
		m.viewportOffset = m.cursor - visibleRows + 1
	}

	// Scroll up if cursor is above viewport
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}

	// Ensure viewport doesn't go beyond list bounds
	maxOffset := len(m.filteredItems) - visibleRows
	if m.viewportOffset > maxOffset {
		m.viewportOffset = Max(0, maxOffset)
	}
}

// getVisibleRows calculates the number of rows that can be displayed
func (m *SelectorModel) getVisibleRows() int {
	if m.maxVisibleRows > 0 {
		return m.maxVisibleRows
	}

	// Calculate based on terminal height
	// Reserve space for: search box (3 lines), help text (2 lines), border/padding (4 lines)
	reservedLines := 9
	availableHeight := m.height - reservedLines

	// Ensure at least 5 items are visible
	if availableHeight < 5 {
		availableHeight = 5
	}

	return availableHeight
}

// Update handles messages for the selector model
func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update width and height for terminal size
		if msg.Width > 64 {
			m.width = 60
		} else {
			m.width = msg.Width - 4
		}
		m.height = msg.Height
		m.adjustViewport() // Recalculate viewport with new height
		return m, nil

	case tea.KeyMsg:
		// Handle search mode key presses
		if m.searchMode {
			switch msg.Type {
			case tea.KeyEsc:
				// Exit search mode
				m.searchMode = false
				return m, nil
			case tea.KeyEnter:
				// Exit search mode and process selection
				m.searchMode = false
				if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
					m.choice = m.filteredItems[m.cursor].Value()
					m.selected = true
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterItems()
				}
				return m, nil
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
					m.adjustViewport()
				}
				return m, nil
			case tea.KeyDown:
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
					m.adjustViewport()
				}
				return m, nil
			default:
				// Add character to search query
				if msg.Type == tea.KeyRunes {
					m.searchQuery += string(msg.Runes)
					m.filterItems()
				}
				return m, nil
			}
		} else {
			// Handle normal navigation mode
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.quitting = true
				return m, tea.Quit

			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
					m.adjustViewport()
				}
				return m, nil

			case tea.KeyDown:
				if m.cursor < len(m.filteredItems)-1 {
					m.cursor++
					m.adjustViewport()
				}
				return m, nil

			case tea.KeyEnter:
				if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
					m.choice = m.filteredItems[m.cursor].Value()
					m.selected = true
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil

			case tea.KeyRunes:
				// Start search mode when typing
				if len(msg.Runes) > 0 {
					if msg.Runes[0] == '/' {
						// Start search mode with '/' key
						m.searchMode = true
						return m, nil
					}
					// Start search with typed character
					m.searchMode = true
					m.searchQuery = string(msg.Runes)
					m.filterItems()
				}
				return m, nil
			}
		}
	}

	return m, nil
}

// View renders the selector
func (m SelectorModel) View() string {
	if m.quitting {
		if m.selected {
			// Find the selected item from filtered items
			if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
				selectedItem := m.filteredItems[m.cursor]
				successStyle := lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
				return successStyle.Render(fmt.Sprintf("✓ Selected: %s", selectedItem.Title()))
			}
		}
		return ""
	}

	var b strings.Builder

	// Search box
	searchBoxStyle := lipgloss.NewStyle().
		Foreground(styles.TextColor).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Width(m.width - 2) // full-width minus padding

	searchPrompt := ">"
	if m.searchMode {
		searchBoxStyle = searchBoxStyle.BorderForeground(styles.PrimaryColor)
	}

	searchText := m.searchQuery
	if searchText == "" && !m.searchMode {
		searchText = "Type press / to search..."
		searchBoxStyle = searchBoxStyle.Foreground(styles.SubtleColor)
	}

	if m.searchMode {
		searchText = searchText + "█" // Add cursor
	}

	b.WriteString(searchBoxStyle.Render(fmt.Sprintf("%s %s", searchPrompt, searchText)))
	b.WriteString("\n")

	// Render filtered items
	if len(m.filteredItems) == 0 {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Align(lipgloss.Center).
			Padding(2, 0)
		b.WriteString(noResultsStyle.Render("No matches found"))
		b.WriteString("\n")
	} else {
		visibleRows := m.getVisibleRows()
		startIdx := m.viewportOffset
		endIdx := Min(startIdx+visibleRows, len(m.filteredItems))

		// Show scroll indicator at top if there are items above
		if startIdx > 0 {
			scrollIndicatorStyle := lipgloss.NewStyle().
				Foreground(styles.SubtleColor).
				Italic(true)
			b.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more above...", startIdx)))
			b.WriteString("\n")
		}

		// Render visible items only
		for i := startIdx; i < endIdx; i++ {
			item := m.filteredItems[i]
			var itemStyle lipgloss.Style
			prefix := "  "

			if i == m.cursor {
				// Selected item
				itemStyle = lipgloss.NewStyle().
					Foreground(styles.PrimaryColor).
					Bold(true)
				prefix = "▶ "
			} else {
				// Normal item
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

		// Show scroll indicator at bottom if there are items below
		if endIdx < len(m.filteredItems) {
			scrollIndicatorStyle := lipgloss.NewStyle().
				Foreground(styles.SubtleColor).
				Italic(true)
			remainingItems := len(m.filteredItems) - endIdx
			b.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more below...", remainingItems)))
			b.WriteString("\n")
		}
	}

	// Show filtered results count if searching
	if m.searchQuery != "" {
		countStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Italic(true).
			Margin(1, 0, 0, 0)
		b.WriteString(countStyle.Render(fmt.Sprintf("Found %02d match(es)", len(m.filteredItems))))
		b.WriteString("\n")
	}

	// Instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Italic(true).
		Margin(1, 0, 0, 0)

	helpText := "Use ↑/↓ to navigate, Enter to select, Esc to cancel, / to search"
	if m.searchMode {
		helpText = "Type to filter, ↑/↓ to navigate, Enter to select, Esc to exit search"
	}
	b.WriteString(helpStyle.Render(helpText))

	return BorderStyle.Render(b.String())
}

// Choice returns the selected choice value
func (m SelectorModel) Choice() string {
	return m.choice
}

// Selected returns whether a choice was made
func (m SelectorModel) Selected() bool {
	return m.selected
}

// High-level selector functions

// SelectFromOptions shows a selector with simple string options
func SelectFromOptions(title string, options []string) (string, error) {
	items := make([]SelectorItem, len(options))
	for i, option := range options {
		items[i] = SelectorItem{
			title: option,
			value: option,
		}
	}

	return SelectFromItems(title, items)
}

// SelectFromOptionsWithMaxRows shows a selector with simple string options and a specific max visible rows
func SelectFromOptionsWithMaxRows(title string, options []string, maxVisibleRows int) (string, error) {
	items := make([]SelectorItem, len(options))
	for i, option := range options {
		items[i] = SelectorItem{
			title: option,
			value: option,
		}
	}

	return SelectFromItemsWithMaxRows(title, items, maxVisibleRows)
}

// SelectFromItems shows a selector with SelectorItem structs
func SelectFromItems(title string, items []SelectorItem) (string, error) {
	model := NewSelectorModel(title, items)

	// Run inline without full-screen mode
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

// SelectFromItemsWithMaxRows shows a selector with SelectorItem structs and a specific max visible rows
func SelectFromItemsWithMaxRows(title string, items []SelectorItem, maxVisibleRows int) (string, error) {
	model := NewSelectorModelWithMaxRows(title, items, maxVisibleRows)

	// Run inline without full-screen mode
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

// SelectOrg shows an organization selector with evaluation journey support
func SelectOrg(orgs []OrgOption) (string, error) {
	items := make([]SelectorItem, len(orgs))
	maxOrgNameWidth := 0
	for _, org := range orgs {
		if len(org.Name) > maxOrgNameWidth {
			maxOrgNameWidth = len(org.Name)
		}
	}
	for i, org := range orgs {
		title := fmt.Sprintf("%s%s", org.Name, strings.Repeat(" ", maxOrgNameWidth-len(org.Name)))
		description := styles.TextDim.Render(fmt.Sprintf("ID: %s", org.ID))

		// Add evaluation journey indicators
		if org.IsEvaluation {
			title = fmt.Sprintf("🚀 %s (Evaluation)", org.Name)
			description = fmt.Sprintf("ID: %s • Perfect for trying out Nuon", org.ID)
		}

		items[i] = SelectorItem{
			title:        title,
			description:  description,
			value:        org.ID,
			isEvaluation: org.IsEvaluation,
		}
	}

	title := "Select an organization"
	if hasEvaluationOrgs(orgs) {
		title = "Select an organization (🚀 = Evaluation mode)"
	}

	return SelectFromItems(title, items)
}

// SelectApp shows an application selector
func SelectApp(apps []AppOption) (string, error) {
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

	return SelectFromItems("Select an application", items)
}

// SelectInstall shows an installation selector
func SelectInstall(installs []InstallOption) (string, error) {
	items := make([]SelectorItem, len(installs))
	// get some widths for padding
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

	return SelectFromItems("Select an installation", items)
}

// SelectWorkflow shows an installation selector
func SelectWorkflow(workflows []WorkflowOption) (string, error) {
	items := make([]SelectorItem, len(workflows))
	// get some widths for padding
	maxWorkflowNameWidth := 0
	for _, workflow := range workflows {
		if len(workflow.Name) > maxWorkflowNameWidth {
			maxWorkflowNameWidth = len(workflow.Name)
		}
	}

	for i, workflow := range workflows {
		items[i] = SelectorItem{
			title:       fmt.Sprintf("%s%s", workflow.Name, strings.Repeat(" ", maxWorkflowNameWidth-len(workflow.Name))),
			description: styles.TextDim.Render(fmt.Sprintf("ID: %s", workflow.ID)),
			value:       workflow.ID,
		}
	}

	return SelectFromItems("Select an workflow", items)
}

// Helper types for the selector functions
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
	ID   string
	Name string
}

// Helper functions
func hasEvaluationOrgs(orgs []OrgOption) bool {
	for _, org := range orgs {
		if org.IsEvaluation {
			return true
		}
	}
	return false
}

// ParseOrgSelection parses a "Name: ID" formatted string (for backward compatibility)
func ParseOrgSelection(selection string) (name, id string) {
	parts := strings.Split(selection, ":")
	if len(parts) >= 2 {
		name = strings.TrimSpace(parts[0])
		id = strings.TrimSpace(parts[1])
	}
	return
}
