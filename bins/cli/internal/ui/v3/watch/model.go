package watch

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	minRequiredWidth  int = 100
	minRequiredHeight int = 20
)

// Exit codes for watch TUI - matches workflows_watch.go constants
const (
	ExitCodeSuccess   = 0
	ExitCodeFailed    = 1
	ExitCodeInterrupt = 130
)

type model struct {
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	installID string

	width  int
	height int

	workflows        []*models.AppWorkflow
	selectedWorkflow *models.AppWorkflow
	selectedIndex    int

	header         viewport.Model
	workflowsList  list.Model
	workflowDetail viewport.Model
	footer         viewport.Model
	focus          string

	help     help.Model
	spinner  spinner.Model
	progress progress.Model
	status   common.StatusBarRequest

	keys keyMap

	error    error
	quitting bool
	loading  bool
	exitCode int
}

func initialWorkflowsList() list.Model {
	workflowsList := list.New([]list.Item{}, list.NewDefaultDelegate(), minRequiredWidth, 0)
	workflowsList.SetShowPagination(true)
	workflowsList.SetShowStatusBar(false)
	workflowsList.SetShowHelp(false)
	workflowsList.SetShowTitle(false)
	return workflowsList
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)
	workflowsList := initialWorkflowsList()
	progress := progress.New()

	m := model{
		ctx:       ctx,
		cfg:       cfg,
		api:       api,
		installID: installID,

		header:         viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(2)),
		workflowsList:  workflowsList,
		workflowDetail: viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(30)),
		footer:         viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(4)),
		focus:          "list",

		help:     help.New(),
		spinner:  s,
		progress: progress,
		status:   common.StatusBarRequest{Message: ""},

		keys: keys,
	}
	m.workflowDetail.SetContent("Select a workflow to view details")

	return m
}

func (m *model) setLogMessage(message string, level string) {
	m.status.Message = message
	m.status.Level = level
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchWorkflowsCmd,
		common.TickCmd(common.DefaultRefreshInterval),
		m.spinner.Tick,
	)
}

func (m *model) resetSelected() {
	m.selectedWorkflow = nil
	m.selectedIndex = -1

	m.keys.Esc.SetHelp("esc", "quit")
	m.keys.OpenWorkflow.SetEnabled(false)

	m.populateWorkflowDetailView()
	m.focus = "list"
}

func (m *model) setSelected() {
	items := m.workflowsList.Items()
	if len(items) == 0 {
		return
	}
	m.selectedIndex = m.workflowsList.Index()

	item := items[m.workflowsList.Index()]
	m.selectedWorkflow = item.(listWorkflow).Workflow()

	m.keys.Esc.SetHelp("esc", "back")
	m.keys.OpenWorkflow.SetEnabled(true)

	m.populateWorkflowDetailView()
}

func (m *model) resize() {
	if m.width == 0 || m.height == 0 {
		return
	}

	headerHeight := 4
	footerHeight := 4
	contentHeight := m.height - headerHeight - footerHeight

	listWidth := int(float64(m.width) * 0.4)
	// Width(listWidth) is total outer width including borders (2) and padding (1 right),
	// so the list content area is listWidth - 3.
	listContentWidth := listWidth - 3
	detailWidth := m.width - listWidth - 2

	m.header.SetWidth(m.width - 2)
	m.header.SetHeight(headerHeight)

	m.workflowsList.SetSize(listContentWidth, contentHeight-2)
	m.workflowDetail.SetWidth(detailWidth)
	m.workflowDetail.SetHeight(contentHeight - 2)

	m.footer.SetWidth(m.width - 2)
	m.footer.SetHeight(footerHeight)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.resize()
}

func (m *model) toggleFocus() {
	if m.focus == "list" {
		m.focus = "detail"
	} else {
		m.focus = "list"
	}
}

func (m *model) handleNav(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == "detail" {
		m.workflowDetail, cmd = m.workflowDetail.Update(msg)
	} else {
		m.workflowsList, cmd = m.workflowsList.Update(msg)
	}
	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case common.TickMsg:
		return m, tea.Batch(
			m.fetchWorkflowsCmd,
			common.TickCmd(common.DefaultRefreshInterval),
		)

	case workflowsFetchedMsg:
		m.handleWorkflowsFetched(msg)

	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			m.exitCode = ExitCodeInterrupt
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Esc):
			if m.selectedWorkflow != nil {
				m.resetSelected()
			} else {
				m.quitting = true
				m.exitCode = ExitCodeSuccess
				return m, tea.Quit
			}

		case key.Matches(msg, m.keys.Up):
			m, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			m, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Left):
			m.toggleFocus()
		case key.Matches(msg, m.keys.Right):
			m.toggleFocus()

		case key.Matches(msg, m.keys.PageDown):
			m.workflowDetail, cmd = m.workflowDetail.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keys.PageUp):
			m.workflowDetail, cmd = m.workflowDetail.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Enter):
			m.setSelected()
			m.workflowsList.Update(msg)

		case key.Matches(msg, m.keys.Tab):
			m.toggleFocus()

		case key.Matches(msg, m.keys.OpenWorkflow):
			m.openWorkflowInBrowser()

		case key.Matches(msg, m.keys.Refresh):
			cmds = append(cmds, m.fetchWorkflowsCmd)
		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func (m model) viewContent() string {
	if m.quitting {
		return "quitting " + m.spinner.View()
	}
	if m.width == 0 {
		return ""
	} else if m.width < minRequiredWidth || m.height < minRequiredHeight {
		content := common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Level:   "warning",
			Content: lipgloss.JoinVertical(
				lipgloss.Center,
				"  This screen is too small, please increase the width.  ",
				fmt.Sprintf("Minimum dimensions %d x %d.  ", minRequiredWidth, minRequiredHeight),
			),
		})
		return content
	}

	header := m.headerView()
	content := ""
	if len(m.workflows) == 0 && m.loading {
		if m.error != nil {
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.workflowDetail.Height(),
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(m.error.Error()),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.workflowDetail.Height(),
				Padding: 1,
				Content: "  Loading  ",
				Level:   "info",
			})
		}
	} else {
		listWidth := int(float64(m.width) * 0.4)
		workflowsList := ""
		if m.focus == "list" {
			workflowsList = appStyleFocus.Width(listWidth).Padding(0, 1, 0, 0).Render(m.workflowsList.View())
		} else {
			workflowsList = appStyleBlur.Width(listWidth).Padding(0, 1, 0, 0).Render(m.workflowsList.View())
		}
		workflowDetail := ""
		if m.focus == "detail" {
			workflowDetail = appStyleFocus.Render(m.workflowDetail.View())
		} else {
			workflowDetail = appStyleBlur.Render(m.workflowDetail.View())
		}
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			workflowsList,
			workflowDetail,
		)
	}
	footer := m.footerView()
	s := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	return s
}

func WatchApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
) int {
	if !cfg.Interactive {
		return watchPlainText(ctx, api, installID)
	}

	m := initialModel(ctx, cfg, api, installID)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		return ExitCodeFailed
	}

	// Extract exit code from final model
	if fm, ok := finalModel.(model); ok {
		return fm.exitCode
	}
	return ExitCodeSuccess
}

// watchPlainText is the non-interactive fallback for WatchApp. It polls
// workflows and prints a plain-text summary, similar to WorkflowsWatch.
func watchPlainText(ctx context.Context, api nuon.Client, installID string) int {
	workflows, _, err := api.GetWorkflows(ctx, installID, nil)
	if err != nil {
		fmt.Printf("Error fetching workflows: %v\n", err)
		return ExitCodeFailed
	}

	if len(workflows) == 0 {
		fmt.Println("No workflows found for this install.")
		return ExitCodeSuccess
	}

	fmt.Printf("Workflows for install %s:\n\n", installID)
	for _, wf := range workflows {
		status := ""
		if wf.Status != nil {
			status = string(wf.Status.Status)
		}
		fmt.Printf("  %s  %-20s  %-12s  %s\n", wf.ID, wf.Name, status, string(wf.Type))
	}
	return ExitCodeSuccess
}
