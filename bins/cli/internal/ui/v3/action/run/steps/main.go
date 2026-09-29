package steps

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	log *common.Logger

	ctx context.Context
	api nuon.Client

	width  int
	height int
	run    *models.AppInstallActionWorkflowRun

	steps          []stepItem
	logsByStep     map[string][]*models.AppOtelLogRecord
	logStream      *models.AppLogStream
	logsCursor     string
	loadingLogs    bool
	logsFetchError error

	selectedStepIndex int
	expandedStepIndex int
	stepsViewport     viewport.Model
	logsViewport      viewport.Model

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
	log, _ := common.NewLogger("run-steps")

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
		stepsViewport:     viewport.New(viewport.WithWidth(width), viewport.WithHeight(height)),
		logsViewport:      viewport.New(viewport.WithWidth(width), viewport.WithHeight(height-headerHeight)),

		help: help.New(),
		keys: keys,
	}

	m.updateStepItems()

	if run != nil && run.LogStream != nil {
		m.logStream = run.LogStream
	}

	return m
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 3

	m.stepsViewport.SetWidth(width)
	m.stepsViewport.SetHeight(height)
	m.logsViewport.SetWidth(width)
	m.logsViewport.SetHeight(height - headerHeight)

	m.setContent()
}

func (m *Model) SetRun(run *models.AppInstallActionWorkflowRun) {
	m.run = run
	m.updateStepItems()

	if run != nil && run.LogStream != nil {
		m.logStream = run.LogStream
	}

	m.setContent()
}

func (m *Model) updateStepItems() {
	if m.run == nil || m.run.Config == nil || m.run.Config.Steps == nil {
		return
	}

	steps := []stepItem{}
	for _, configStep := range m.run.Config.Steps {
		if configStep == nil {
			continue
		}

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

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		common.TickCmd(common.DefaultRefreshInterval),
		m.fetchLogsCmd,
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case common.TickMsg:
		return m, tea.Batch(
			m.fetchLogsCmd,
			common.TickCmd(common.DefaultRefreshInterval),
		)

	case logsFetchedMsg:
		m.handleLogsFetched(msg)

	case tea.KeyPressMsg:
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
				m.logsViewport, cmd = m.logsViewport.Update(msg)
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "up", "k":
				m.moveStepSelection(-1)
				m.setContent()
				m.stepsViewport, cmd = m.stepsViewport.Update(msg)
				cmds = append(cmds, cmd)
			case "down", "j":
				m.moveStepSelection(1)
				m.setContent()
				m.stepsViewport, cmd = m.stepsViewport.Update(msg)
				cmds = append(cmds, cmd)
			case "enter":
				m.toggleStepExpansion()
				if m.expandedStepIndex >= 0 {
					cmds = append(cmds, m.fetchLogsCmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if m.expandedStepIndex >= 0 && m.expandedStepIndex < len(m.steps) {
		item := m.steps[m.expandedStepIndex]
		status := item.getStatus()
		name := item.getName()
		duration := item.getExecutionDuration()

		statusStyle := styles.GetStatusStyle(models.AppStatus(status))
		statusText := statusStyle.Render(fmt.Sprintf("[%s] ", status))
		durationString := styles.TextSubtle.Render(duration)
		spacer := strings.Repeat(" ", (m.width-6)-(lipgloss.Width(statusText)+lipgloss.Width(name)+lipgloss.Width(durationString)))
		header := lipgloss.NewStyle().Padding(1).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				statusText,
				name,
				spacer,
				durationString,
			),
		)

		return lipgloss.JoinVertical(lipgloss.Top, header, m.logsViewport.View())
	}

	return m.stepsViewport.View()
}

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

func (m *Model) toggleStepExpansion() {
	if m.expandedStepIndex == m.selectedStepIndex {
		m.expandedStepIndex = -1
	} else {
		m.expandedStepIndex = m.selectedStepIndex
	}
	m.setContent()
}

func (m Model) renderStepItem(index int, item stepItem, selected bool) string {
	status := item.getStatus()
	name := item.getName()
	duration := item.getExecutionDuration()

	stepStyle := styles.GetStepStyle(status, selected)

	statusStyle := styles.GetStatusStyle(models.AppStatus(status))
	statusText := statusStyle.Render(fmt.Sprintf("[%s] ", status))

	durationString := styles.TextSubtle.Render(duration)
	spacer := strings.Repeat(" ", (m.width-6)-(lipgloss.Width(statusText)+lipgloss.Width(name)+lipgloss.Width(durationString)))
	content := lipgloss.JoinHorizontal(lipgloss.Left,
		statusText,
		name,
		spacer,
		durationString,
	)
	return stepStyle.Padding(1).Width(m.width - 2).Render(content)
}

func (m *Model) setContent() {
	if m.expandedStepIndex >= 0 && m.expandedStepIndex < len(m.steps) {
		item := m.steps[m.expandedStepIndex]

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

		m.logsViewport.SetContent(logsContent)

	} else {
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

func (m Model) getStepLogs(item stepItem) string {
	maxLength := m.width - 6
	stepName := item.getName()
	logs, ok := m.logsByStep[stepName]
	if !ok || len(logs) == 0 {
		return ""
	}

	var logLines []string
	for _, log := range logs {
		severity := strings.ToUpper(log.SeverityText)

		timestamp := log.Timestamp
		if len(timestamp) > 19 {
			timestamp = timestamp[:19]
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

func getLevelStyle(level string) lipgloss.Style {
	switch level {
	case "ERROR", "FATAL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	case "WARN", "WARNING":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	case "DEBUG", "TRACE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	}
}
