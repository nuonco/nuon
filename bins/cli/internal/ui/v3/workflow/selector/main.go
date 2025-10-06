package selector

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

const (
	limit = 20
)

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Next     key.Binding
	Previous key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Enter, k.Next, k.Previous, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Next, k.Previous},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Next: key.NewBinding(
		key.WithKeys("n", "right", "l"),
		key.WithHelp("n/→", "next page"),
	),
	Previous: key.NewBinding(
		key.WithKeys("p", "left", "h"),
		key.WithHelp("p/←", "prev page"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↳", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

type model struct {
	ctx       context.Context
	cfg       *config.Config
	api       nuon.Client
	installID string

	workflows  []*models.AppWorkflow
	table      table.Model
	spinner    spinner.Model
	help       help.Model
	keys       keyMap
	loading    bool
	offset     int
	hasMore    bool
	selectedID string
	width      int
	height     int
	err        error
	quitting   bool
}

type workflowsLoadedMsg struct {
	workflows []*models.AppWorkflow
	hasMore   bool
	err       error
}

func loadWorkflows(ctx context.Context, api nuon.Client, installID string, offset int) tea.Cmd {
	return func() tea.Msg {
		workflows, hasMore, err := api.GetWorkflows(ctx, installID, &models.GetPaginatedQuery{
			Offset: offset,
			Limit:  limit,
		})
		return workflowsLoadedMsg{
			workflows: workflows,
			hasMore:   hasMore,
			err:       err,
		}
	}
}

func initialModel(ctx context.Context, cfg *config.Config, api nuon.Client, installID string) model {
	columns := []table.Column{
		{Title: "ID", Width: 28},
		{Title: "NAME", Width: 20},
		{Title: "TYPE", Width: 15},
		{Title: "STATUS", Width: 15},
		{Title: "STARTED AT", Width: 20},
		{Title: "FINISHED AT", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.PrimaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(styles.PrimaryColor)
	s.Selected = s.Selected.
		Foreground(styles.TextColor).
		Background(styles.PrimaryColor).
		Bold(false)
	t.SetStyles(s)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	return model{
		ctx:       ctx,
		cfg:       cfg,
		api:       api,
		installID: installID,
		table:     t,
		spinner:   sp,
		help:      help.New(),
		keys:      keys,
		loading:   true,
		offset:    0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadWorkflows(m.ctx, m.api, m.installID, m.offset),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case workflowsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.workflows = msg.workflows
		m.hasMore = msg.hasMore

		// Convert workflows to table rows
		rows := []table.Row{}
		for _, workflow := range m.workflows {
			startedAt := ""
			if workflow.StartedAt != "" {
				if t, err := time.Parse(time.RFC3339Nano, workflow.StartedAt); err == nil {
					startedAt = t.Format(time.Stamp)
				}
			}
			finishedAt := ""
			if workflow.FinishedAt != "" {
				if t, err := time.Parse(time.RFC3339Nano, workflow.FinishedAt); err == nil {
					finishedAt = t.Format(time.Stamp)
				}
			}
			status := ""
			if workflow.Status != nil {
				status = string(workflow.Status.Status)
			}

			rows = append(rows, table.Row{
				workflow.ID,
				workflow.Name,
				string(workflow.Type),
				status,
				startedAt,
				finishedAt,
			})
		}
		m.table.SetRows(rows)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			if len(m.workflows) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.workflows) {
					m.selectedID = m.workflows[selectedIdx].ID
					return m, tea.Quit
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Next):
			if m.hasMore && !m.loading {
				m.offset += limit
				m.loading = true
				return m, tea.Batch(
					loadWorkflows(m.ctx, m.api, m.installID, m.offset),
					m.spinner.Tick,
				)
			}
			return m, nil

		case key.Matches(msg, m.keys.Previous):
			if m.offset > 0 && !m.loading {
				m.offset -= limit
				if m.offset < 0 {
					m.offset = 0
				}
				m.loading = true
				return m, tea.Batch(
					loadWorkflows(m.ctx, m.api, m.installID, m.offset),
					m.spinner.Tick,
				)
			}
			return m, nil

		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}

	default:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		if m.selectedID != "" {
			return ""
		}
		return "Cancelled.\n"
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(styles.ErrorColor).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ErrorColor)
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	var content string

	if m.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(styles.AccentColor).
			Padding(1, 2)
		content = loadingStyle.Render(fmt.Sprintf("%s Loading workflows...", m.spinner.View()))
	} else if len(m.workflows) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Padding(1, 2)
		content = emptyStyle.Render("No workflows found")
	} else {
		content = m.table.View()
	}

	pageInfo := ""
	if !m.loading {
		pageStyle := lipgloss.NewStyle().
			Foreground(styles.SubtleColor).
			Padding(0, 1)
		start := m.offset + 1
		end := m.offset + len(m.workflows)
		moreIndicator := ""
		if m.hasMore {
			moreIndicator = "+"
		}
		pageInfo = pageStyle.Margin(1, 0).Render(fmt.Sprintf("Showing %d-%d%s (offset: %d)", start, end, moreIndicator, m.offset))
	}

	helpView := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Padding(1, 1).
		Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		pageInfo,
		helpView,
	)
}

// WorkflowSelectorApp runs the workflow selector and returns the selected workflow ID
func WorkflowSelectorApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
) (string, error) {
	m := initialModel(ctx, cfg, api, installID)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running workflow selector: %w", err)
	}

	if fm, ok := finalModel.(model); ok {
		if fm.selectedID != "" {
			return fm.selectedID, nil
		}
	}

	return "", fmt.Errorf("no workflow selected")
}
