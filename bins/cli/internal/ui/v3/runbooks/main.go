package runbooks

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

// View states
type viewState int

const (
	viewLoading viewState = iota
	viewList
	viewDetail
	viewRunConfirm
	viewRunning
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	statusStyles = map[string]lipgloss.Style{
		"queued":      lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")),
		"in-progress": lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")),
		"finished":    lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")),
		"error":       lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")),
		"cancelled":   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")),
		"unknown":     dimStyle,
	}

	readmeStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))

	confirmStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#FFD700"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF7F"))

	errorMsgStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4500"))
)

// Key bindings
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Run     key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Yes     key.Binding
	No      key.Binding
	Help    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/view"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Run: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "run runbook"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "refresh"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Run, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Run, k.Refresh, k.Back},
		{k.Quit, k.Help},
	}
}

// Messages
type runbooksLoadedMsg struct {
	runbooks []*nuon.InstallRunbook
}

type runbookDetailMsg struct {
	runbook *nuon.InstallRunbook
}

type runTriggeredMsg struct {
	run *nuon.InstallRunbookRun
}

type errMsg struct {
	err error
}

type tickMsg struct{}

// Model
type model struct {
	log *common.Logger
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	installID string

	state    viewState
	runbooks []*nuon.InstallRunbook
	cursor   int

	// detail view
	selectedRunbook *nuon.InstallRunbook

	// run state
	lastRun  *nuon.InstallRunbookRun
	errorMsg string

	// ui components
	spinner  spinner.Model
	help     help.Model
	keys     keyMap
	width    int
	height   int
	showHelp bool
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
) model {
	log, _ := common.NewLogger("runbooks")
	s := spinner.New()
	s.Spinner = spinner.Dot

	return model{
		log:       log,
		ctx:       ctx,
		cfg:       cfg,
		api:       api,
		installID: installID,
		state:     viewLoading,
		spinner:   s,
		help:      help.New(),
		keys:      keys,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadRunbooks())
}

func (m model) loadRunbooks() tea.Cmd {
	return func() tea.Msg {
		runbooks, err := m.api.GetInstallRunbooks(m.ctx, m.installID)
		if err != nil {
			return errMsg{err: err}
		}
		return runbooksLoadedMsg{runbooks: runbooks}
	}
}

func (m model) loadRunbookDetail(runbookID string) tea.Cmd {
	return func() tea.Msg {
		runbook, err := m.api.GetInstallRunbook(m.ctx, m.installID, runbookID)
		if err != nil {
			return errMsg{err: err}
		}
		return runbookDetailMsg{runbook: runbook}
	}
}

func (m model) triggerRun(runbookID string) tea.Cmd {
	return func() tea.Msg {
		run, err := m.api.CreateInstallRunbookRun(m.ctx, m.installID, runbookID)
		if err != nil {
			return errMsg{err: err}
		}
		return runTriggeredMsg{run: run}
	}
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case errMsg:
		m.errorMsg = msg.err.Error()
		if m.state == viewLoading {
			m.state = viewList
		}

	case runbooksLoadedMsg:
		m.runbooks = msg.runbooks
		m.state = viewList
		m.errorMsg = ""

	case runbookDetailMsg:
		m.selectedRunbook = msg.runbook
		m.state = viewDetail
		m.errorMsg = ""

	case runTriggeredMsg:
		m.lastRun = msg.run
		m.state = viewRunning
		m.errorMsg = ""
		cmds = append(cmds, m.tickCmd())

	case tickMsg:
		if m.state == viewRunning && m.selectedRunbook != nil {
			cmds = append(cmds, m.loadRunbookDetail(m.selectedRunbook.RunbookID))
			cmds = append(cmds, m.tickCmd())
		}

	case tea.KeyPressMsg:
		switch m.state {
		case viewList:
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.runbooks)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Enter):
				if len(m.runbooks) > 0 {
					rb := m.runbooks[m.cursor]
					cmds = append(cmds, m.loadRunbookDetail(rb.RunbookID))
					m.state = viewLoading
				}
			case key.Matches(msg, m.keys.Run):
				if len(m.runbooks) > 0 {
					m.selectedRunbook = m.runbooks[m.cursor]
					m.state = viewRunConfirm
				}
			case key.Matches(msg, m.keys.Refresh):
				m.state = viewLoading
				cmds = append(cmds, m.loadRunbooks())
			case key.Matches(msg, m.keys.Help):
				m.showHelp = !m.showHelp
			}

		case viewDetail:
			switch {
			case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Quit):
				m.state = viewList
				m.selectedRunbook = nil
			case key.Matches(msg, m.keys.Run):
				if m.selectedRunbook != nil {
					m.state = viewRunConfirm
				}
			}

		case viewRunConfirm:
			switch {
			case key.Matches(msg, m.keys.Yes):
				if m.selectedRunbook != nil {
					m.state = viewLoading
					cmds = append(cmds, m.triggerRun(m.selectedRunbook.RunbookID))
				}
			case key.Matches(msg, m.keys.No), key.Matches(msg, m.keys.Back):
				if m.selectedRunbook != nil {
					m.state = viewDetail
				} else {
					m.state = viewList
				}
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}

		case viewRunning:
			switch {
			case key.Matches(msg, m.keys.Back):
				m.state = viewDetail
				m.lastRun = nil
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
		}

	default:
		var cmd tea.Cmd
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
	var b strings.Builder

	// Header
	header := titleStyle.Render(fmt.Sprintf(" Runbooks - Install %s ", truncateID(m.installID)))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorMsgStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n\n")
	}

	switch m.state {
	case viewLoading:
		b.WriteString(m.spinner.View() + " Loading...")

	case viewList:
		b.WriteString(m.viewList())

	case viewDetail:
		b.WriteString(m.viewDetailContent())

	case viewRunConfirm:
		b.WriteString(m.viewRunConfirm())

	case viewRunning:
		b.WriteString(m.viewRunStatus())
	}

	// Footer with help
	b.WriteString("\n\n")
	if m.showHelp {
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString(dimStyle.Render("? help"))
	}

	return b.String()
}

func (m model) viewList() string {
	if len(m.runbooks) == 0 {
		return dimStyle.Render("No runbooks found for this install.")
	}

	var b strings.Builder
	b.WriteString(dimStyle.Render(fmt.Sprintf("%d runbook(s) available", len(m.runbooks))))
	b.WriteString("\n\n")

	for i, rb := range m.runbooks {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		name := rb.Runbook.Name
		if name == "" {
			name = rb.RunbookID
		}

		status := rb.Status
		if status == "" {
			status = "unknown"
		}
		statusSt, ok := statusStyles[status]
		if !ok {
			statusSt = dimStyle
		}

		// Step count from latest config
		stepCount := 0
		if len(rb.Runbook.Configs) > 0 {
			stepCount = len(rb.Runbook.Configs[0].Steps)
		}

		line := fmt.Sprintf("%s%s  %s  %s",
			cursor,
			style.Render(name),
			statusSt.Render("["+status+"]"),
			dimStyle.Render(fmt.Sprintf("(%d steps)", stepCount)),
		)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("enter: view details  r: run  R: refresh  q: quit"))

	return b.String()
}

func (m model) viewDetailContent() string {
	if m.selectedRunbook == nil {
		return "No runbook selected."
	}

	var b strings.Builder
	rb := m.selectedRunbook

	name := rb.Runbook.Name
	if name == "" {
		name = rb.RunbookID
	}

	b.WriteString(selectedStyle.Render(name))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("ID: " + rb.RunbookID))
	b.WriteString("\n\n")

	// Status
	status := rb.Status
	if status == "" {
		status = "unknown"
	}
	statusSt, ok := statusStyles[status]
	if !ok {
		statusSt = dimStyle
	}
	b.WriteString(fmt.Sprintf("Status: %s", statusSt.Render(status)))
	b.WriteString("\n\n")

	// Steps
	if len(rb.Runbook.Configs) > 0 {
		cfg := rb.Runbook.Configs[0]

		if len(cfg.Steps) > 0 {
			b.WriteString(normalStyle.Render("Steps:"))
			b.WriteString("\n")
			for _, step := range cfg.Steps {
				stepType := step.Type
				if stepType == "" {
					stepType = "unknown"
				}
				detail := ""
				switch stepType {
				case "deploy":
					if step.ComponentName != "" {
						detail = " -> " + step.ComponentName
					}
				case "action":
					if step.ActionWorkflowID != "" {
						detail = " -> " + step.ActionWorkflowID
					}
					if step.Command != "" {
						detail = " -> " + step.Command
					}
				}
				b.WriteString(fmt.Sprintf("  %d. %s [%s]%s\n",
					step.Idx+1,
					step.Name,
					dimStyle.Render(stepType),
					dimStyle.Render(detail),
				))
			}
			b.WriteString("\n")
		}

		// Readme
		if cfg.Readme != "" {
			b.WriteString(normalStyle.Render("Readme:"))
			b.WriteString("\n")
			b.WriteString(readmeStyle.Render(cfg.Readme))
			b.WriteString("\n")
		}
	}

	// Recent runs
	if len(rb.Runs) > 0 {
		b.WriteString("\n")
		b.WriteString(normalStyle.Render("Recent Runs:"))
		b.WriteString("\n")
		for _, run := range rb.Runs {
			runStatus := run.Status
			if runStatus == "" {
				runStatus = "unknown"
			}
			rStatusSt, ok := statusStyles[runStatus]
			if !ok {
				rStatusSt = dimStyle
			}
			b.WriteString(fmt.Sprintf("  %s  %s  %s\n",
				dimStyle.Render(run.ID),
				rStatusSt.Render("["+runStatus+"]"),
				dimStyle.Render(run.CreatedAt.Format(time.RFC3339)),
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("r: run  esc: back  q: quit"))

	return b.String()
}

func (m model) viewRunConfirm() string {
	if m.selectedRunbook == nil {
		return "No runbook selected."
	}

	name := m.selectedRunbook.Runbook.Name
	if name == "" {
		name = m.selectedRunbook.RunbookID
	}

	content := fmt.Sprintf("Run runbook \"%s\"?\n\n%s / %s",
		name,
		selectedStyle.Render("y: yes"),
		dimStyle.Render("n: no"),
	)

	return confirmStyle.Render(content)
}

func (m model) viewRunStatus() string {
	if m.lastRun == nil {
		return m.spinner.View() + " Triggering run..."
	}

	var b strings.Builder

	status := m.lastRun.Status
	if status == "" {
		status = "unknown"
	}
	statusSt, ok := statusStyles[status]
	if !ok {
		statusSt = dimStyle
	}

	b.WriteString(fmt.Sprintf("Run: %s\n", dimStyle.Render(m.lastRun.ID)))
	b.WriteString(fmt.Sprintf("Status: %s\n", statusSt.Render(status)))

	if m.lastRun.StatusDescription != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", m.lastRun.StatusDescription))
	}

	switch status {
	case "finished":
		b.WriteString("\n")
		b.WriteString(successStyle.Render("Runbook completed successfully!"))
	case "error", "cancelled":
		b.WriteString("\n")
		b.WriteString(errorMsgStyle.Render("Runbook " + status))
	default:
		b.WriteString("\n")
		b.WriteString(m.spinner.View() + " Running...")
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("esc: back  q: quit"))

	return b.String()
}

func truncateID(id string) string {
	if len(id) > 16 {
		return id[:16] + "..."
	}
	return id
}

// App is the entry point for the runbooks TUI.
func App(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
) {
	if !cfg.Interactive {
		fmt.Fprintln(os.Stderr, "interactive terminal required; use --json flag for non-interactive output")
		os.Exit(1)
	}

	m := initialModel(ctx, cfg, api, installID)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v", err)
		os.Exit(1)
	}
}
