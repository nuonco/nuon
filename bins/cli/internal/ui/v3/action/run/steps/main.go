package steps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	dataRefreshInterval time.Duration = time.Second * 5
)

// Model represents the steps list and detail view
type Model struct {
	// internal
	log *common.Logger

	// common/base
	ctx context.Context
	api nuon.Client

	// passed down from parent
	width  int
	height int
	run    *models.AppInstallActionWorkflowRun

	// data
	steps          []stepItem
	logsByStep     map[string][]*models.AppOtelLogRecord // keyed by step name
	logStream      *models.AppLogStream
	logsCursor     string
	loadingLogs    bool
	logsFetchError error

	// ui state
	selectedStepIndex int
	expandedStepIndex int
	stepsViewport     viewport.Model // holds the list of steps
	logsViewport      viewport.Model // holds the logs for expanded step

	// ui components
	help help.Model
	keys keyMap
}

func (m Model) ExpandedStepIndex() int {
	return m.expandedStepIndex
}

func New(
	ctx context.Context,
	api nuon.Client,
	width int,
	height int,
	run *models.AppInstallActionWorkflowRun,
) Model {
	// this model should be rendered as large as the width/height defines.
	log, _ := common.NewLogger("run-steps")

	// Calculate header height for expanded view
	headerHeight := 3

	m := Model{
		log:    log,
		ctx:    ctx,
		api:    api,
		width:  width,
		height: height,
		run:    run,

		logsByStep:        make(map[string][]*models.AppOtelLogRecord),
		logsCursor:        "0",
		selectedStepIndex: 0,
		expandedStepIndex: -1,
		stepsViewport:     viewport.New(width, height),
		logsViewport:      viewport.New(width, height-headerHeight),

		help: help.New(),
		keys: keys,
	}

	// Initialize steps from run data
	m.updateStepItems()

	// Initialize log stream if available
	if run != nil && run.LogStream != nil {
		m.logStream = run.LogStream
	}

	return m
}

// SetSize updates the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate header height for expanded view
	headerHeight := 3

	// Update both viewports
	m.stepsViewport.Width = width
	m.stepsViewport.Height = height
	m.logsViewport.Width = width
	m.logsViewport.Height = height - headerHeight

	m.setContent()
}

// SetRun updates the run data
func (m *Model) SetRun(run *models.AppInstallActionWorkflowRun) {
	m.run = run
	m.updateStepItems()

	// Update log stream if available
	if run != nil && run.LogStream != nil {
		m.logStream = run.LogStream
	}

	// Update viewport content with new run data
	m.setContent()
}

// updateStepItems updates the step list based on current run data
func (m *Model) updateStepItems() {
	if m.run == nil || m.run.Config == nil || m.run.Config.Steps == nil {
		return
	}

	steps := []stepItem{}
	for _, configStep := range m.run.Config.Steps {
		if configStep == nil {
			continue
		}

		// Find matching run step
		var runStep *models.AppInstallActionWorkflowRunStep
		for _, rs := range m.run.Steps {
			if rs != nil && rs.StepID == configStep.ID {
				runStep = rs
				break
			}
		}

		steps = append(steps, stepItem{
			configStep: configStep,
			runStep:    runStep,
		})
	}

	m.steps = steps
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tick,
		m.fetchLogsCmd,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Refresh logs periodically
		return m, tea.Batch(
			m.fetchLogsCmd,
			tea.Tick(dataRefreshInterval, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)

	case logsFetchedMsg:
		m.handleLogsFetched(msg)

	case tea.KeyMsg:
		// If a step is expanded, handle logs viewport scrolling or collapsing
		if m.expandedStepIndex >= 0 {
			switch {
			case key.Matches(msg, keys.Enter):
				m.toggleStepExpansion()
				return m, nil
			case key.Matches(msg, keys.Esc):
				if m.expandedStepIndex != -1 {
					m.toggleStepExpansion()
				} else {
					// TODO: send quite message to the parent
					return m, tea.Quit
				}
				return m, nil
			case msg.String() == "up", msg.String() == "k", msg.String() == "down", msg.String() == "j":
				// Scroll the logs viewport (header stays fixed)
				m.logsViewport, cmd = m.logsViewport.Update(msg)
				return m, cmd
			}
		} else {
			// Normal navigation mode - list of steps
			switch msg.String() {
			case "up", "k":
				m.moveStepSelection(-1)
				m.setContent() // Update content to reflect new selection
				// Also update viewport to keep selection visible
				m.stepsViewport, cmd = m.stepsViewport.Update(msg)
				cmds = append(cmds, cmd)
			case "down", "j":
				m.moveStepSelection(1)
				m.setContent() // Update content to reflect new selection
				// Also update viewport to keep selection visible
				m.stepsViewport, cmd = m.stepsViewport.Update(msg)
				cmds = append(cmds, cmd)
			case "enter":
				m.toggleStepExpansion()
				if m.expandedStepIndex >= 0 {
					// Start fetching logs for the expanded step
					cmds = append(cmds, m.fetchLogsCmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Branch 1: Step is expanded - render fixed header + scrollable logs viewport
	if m.expandedStepIndex >= 0 && m.expandedStepIndex < len(m.steps) {
		item := m.steps[m.expandedStepIndex]
		status := item.getStatus()
		name := item.getName()
		duration := item.getExecutionDuration()

		// Build fixed header: [status] name ... duration
		statusStyle := styles.GetStatusStyle(models.AppStatus(status))
		statusText := statusStyle.Render(fmt.Sprintf("[%s] ", status))
		durationString := fmt.Sprintf("%s", styles.TextSubtle.Render(duration))
		spacer := strings.Repeat(" ", (m.width-6)-(lipgloss.Width(statusText)+lipgloss.Width(name)+lipgloss.Width(durationString)))
		header := lipgloss.NewStyle().Padding(1).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				statusText,
				name,
				spacer,
				durationString,
			),
		)

		// Combine fixed header with scrollable logs viewport
		return lipgloss.JoinVertical(lipgloss.Top, header, m.logsViewport.View())
	}

	// Branch 2: No step expanded - show scrollable list of steps
	return m.stepsViewport.View()
}

// moveStepSelection moves the selection up or down
func (m *Model) moveStepSelection(delta int) {
	if len(m.steps) == 0 {
		return
	}

	m.selectedStepIndex += delta
	if m.selectedStepIndex < 0 {
		m.selectedStepIndex = 0
	}
	if m.selectedStepIndex >= len(m.steps) {
		m.selectedStepIndex = len(m.steps) - 1
	}
}

// toggleStepExpansion toggles the expansion of the selected step
func (m *Model) toggleStepExpansion() {
	if m.expandedStepIndex == m.selectedStepIndex {
		// Collapse
		m.expandedStepIndex = -1
	} else {
		// Expand
		m.expandedStepIndex = m.selectedStepIndex
	}
	// Update viewport content based on new expansion state
	m.setContent()
}

// renderStepItem renders a single step item (collapsed view)
func (m Model) renderStepItem(index int, item stepItem, selected bool) string {
	status := item.getStatus()
	name := item.getName()
	duration := item.getExecutionDuration()

	// Get the style based on status and selection
	stepStyle := getStepStyle(status, selected)

	// Build the content
	// icon := getStepStatusIcon(status)
	statusStyle := styles.GetStatusStyle(models.AppStatus(status))
	statusText := statusStyle.Render(fmt.Sprintf("[%s] ", status))

	durationString := fmt.Sprintf("%s", styles.TextSubtle.Render(duration))
	spacer := strings.Repeat(" ", (m.width-6)-(lipgloss.Width(statusText)+lipgloss.Width(name)+lipgloss.Width(durationString)))
	// First line: [status] name ... duration
	content := lipgloss.JoinHorizontal(lipgloss.Left,
		statusText,
		name,
		spacer,
		durationString,
	)
	return stepStyle.Padding(1).Width(m.width - 2).Render(content)
}

// setContent updates the viewport content based on the current state
// This method contains all the branching logic for what content to display
func (m *Model) setContent() {
	// Branch 1: Step is expanded - populate logsViewport only
	if m.expandedStepIndex >= 0 && m.expandedStepIndex < len(m.steps) {
		item := m.steps[m.expandedStepIndex]

		// Get logs content
		logsContent := m.getStepLogs(item)
		if logsContent == "" {
			if m.loadingLogs {
				logsContent = styles.TextSubtle.Italic(true).Padding(1).Render("loading logs...")
			} else if m.logsFetchError != nil {
				logsContent = styles.TextSubtle.Foreground(styles.ErrorColor).Padding(1).Render("Error loading logs: " + m.logsFetchError.Error())
			} else {
				logsContent = styles.TextSubtle.Italic(true).Padding(1).Render("No logs available")
			}
		}

		// Set logs content in logsViewport (no header here)
		m.logsViewport.SetContent(logsContent)

	} else {
		// Branch 2: No step expanded - show list of all steps in stepsViewport
		var content string
		if m.run == nil {
			content = styles.TextSubtle.Italic(true).Padding(1).Render("Loading")
		} else if len(m.steps) == 0 {
			content = styles.TextSubtle.Italic(true).Render("No steps available")
		} else {
			var stepViews []string
			for i, step := range m.steps {
				selected := i == m.selectedStepIndex
				stepViews = append(stepViews, m.renderStepItem(i, step, selected))
			}
			content = lipgloss.JoinVertical(lipgloss.Top, stepViews...)
		}
		m.stepsViewport.SetContent(content)
	}
}

func (m Model) preProcessLog(text string) string {
	maxLength := m.width - 6
	if strings.Contains(text, "\n") {
		text = strings.Split(text, "\n")[0]
	}
	if strings.Contains(text, "\r") {
		text = strings.Split(text, "\r")[0]
	}
	if len(text) > maxLength {
		return text[:maxLength]
	}
	return text
}

// getStepLogs returns the logs for a given step
func (m Model) getStepLogs(item stepItem) string {
	maxLength := m.width - 6
	stepName := item.getName()
	logs, ok := m.logsByStep[stepName]
	if !ok || len(logs) == 0 {
		return ""
	}

	// Format logs for display
	var logLines []string
	for _, log := range logs {
		// Get severity level from log attributes or default to INFO
		severity := strings.ToUpper(log.SeverityText)

		// Format: [SEVERITY] timestamp: body
		timestamp := log.Timestamp
		if len(timestamp) > 19 {
			timestamp = timestamp[:19] // Truncate to readable format
		}

		levelStyle := getLevelStyle(severity)
		line := lipgloss.NewStyle().Width(maxLength).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				levelStyle.Render(fmt.Sprintf("[%s]", severity[0:1])),
				" ",
				styles.TextSubtle.Render(timestamp),
				" ",
				m.preProcessLog(log.Body),
			),
		)
		logLines = append(logLines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, logLines...)
}

// getLevelStyle returns the style for a given log level
func getLevelStyle(level string) lipgloss.Style {
	switch level {
	case "ERROR", "FATAL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	case "WARN", "WARNING":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	case "DEBUG", "TRACE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	default: // INFO
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	}
}
