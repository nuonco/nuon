package detail

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	ac "github.com/nuonco/nuon/bins/cli/internal/ui/v3/action/common"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

const (
	minRequiredWidth  int = 100
	minRequiredHeight int = 20
)

type ViewMode string
type FocusArea string

const (
	ExecuteView    ViewMode  = "execute"
	RunsView       ViewMode  = "runs"
	RunsFocusArea  FocusArea = "runs"
	StepsFocusArea FocusArea = "steps"
)

type Model struct {
	log *common.Logger
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	installID        string
	actionWorkflowID string

	width      int
	height     int
	runsWidth  int
	stepsWidth int

	installActionWorkflow *models.AppInstallActionWorkflow
	latestConfig          *models.AppActionWorkflowConfig

	workflowLoading bool
	configLoading   bool

	header       viewport.Model
	runsList     list.Model
	actionConfig viewport.Model
	footer       viewport.Model
	focus        FocusArea

	spinner spinner.Model

	status common.StatusBarRequest

	help help.Model

	keys keyMap

	viewMode       ViewMode
	formInputs     []textinput.Model
	formFocusIndex int
	formMappings   []executeInputMapping
	formSubmitting bool
	formError      error
	formViewport   viewport.Model

	error    error
	quitting bool
	loading  bool
}

func initialRunsList() list.Model {
	runsList := list.New([]list.Item{}, list.NewDefaultDelegate(), minRequiredWidth, 0)
	runsList.SetShowPagination(false)
	runsList.SetShowStatusBar(false)
	runsList.SetShowHelp(false)
	runsList.SetShowTitle(false)
	return runsList
}

func New(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
) Model {
	m := initialModel(
		ctx,
		cfg,
		api,
		installID,
		actionWorkflowID,
	)
	return m
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
) Model {
	log, _ := common.NewLogger("install-action-detail")
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)
	runsList := initialRunsList()

	m := Model{
		log:              log,
		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		installID:        installID,
		actionWorkflowID: actionWorkflowID,

		header:       viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(2)),
		runsList:     runsList,
		actionConfig: viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(30)),
		footer:       viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(4)),
		focus:        RunsFocusArea,

		help:    help.New(),
		spinner: s,
		status:  common.StatusBarRequest{Message: ""},

		keys:         keys,
		viewMode:     RunsView,
		formViewport: viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(30)),
	}
	m.actionConfig.SetContent("Loading")

	return m
}

func (m *Model) setLogMessage(message string, level string) {
	m.status.Message = message
	m.status.Level = level
}

func (m *Model) initializeExecuteForm() {
	m.formInputs = make([]textinput.Model, 0)
	m.formMappings = make([]executeInputMapping, 0)

	envVarsMap := make(map[string]string)
	if m.latestConfig != nil && m.latestConfig.Steps != nil {
		for _, step := range m.latestConfig.Steps {
			if step == nil || step.EnvVars == nil {
				continue
			}

			for name, value := range step.EnvVars {
				if _, exists := envVarsMap[name]; !exists {
					envVarsMap[name] = value
				}
			}
		}
	}

	var sortedNames []string
	for name := range envVarsMap {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		value := envVarsMap[name]

		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("Enter %s", name)
		ti.CharLimit = 500
		ti.SetWidth(50)
		ti.Prompt = ""

		if value != "" {
			ti.SetValue(value)
		}

		m.formInputs = append(m.formInputs, ti)
		m.formMappings = append(m.formMappings, executeInputMapping{
			name:  name,
			value: value,
			input: name,
		})
	}

	if len(m.formInputs) > 0 {
		m.formInputs[0].Focus()
		m.formFocusIndex = 0
	}

	m.updateFormViewportContent()
}

func (m *Model) updateFormViewportContent() {
	content := m.renderExecuteFormContent()
	m.formViewport.SetContent(content)
}

func (m *Model) nextFormInput() {
	if len(m.formInputs) == 0 {
		return
	}

	if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
		m.formInputs[m.formFocusIndex].Blur()
	}

	m.formFocusIndex++
	if m.formFocusIndex >= len(m.formInputs) {
		m.formFocusIndex = 0
	}

	if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
		m.formInputs[m.formFocusIndex].Focus()
	}
}

func (m *Model) prevFormInput() {
	if len(m.formInputs) == 0 {
		return
	}

	if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
		m.formInputs[m.formFocusIndex].Blur()
	}

	m.formFocusIndex--
	if m.formFocusIndex < 0 {
		m.formFocusIndex = len(m.formInputs) - 1
	}

	if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
		m.formInputs[m.formFocusIndex].Focus()
	}
}

func (m *Model) submitExecuteForm() tea.Cmd {
	return func() tea.Msg {
		envVars := make(map[string]string)
		for i, mapping := range m.formMappings {
			value := strings.TrimSpace(m.formInputs[i].Value())
			if value != "" {
				envVars[mapping.name] = value
			} else if mapping.value != "" {
				envVars[mapping.name] = mapping.value
			}
		}

		configID := m.latestConfig.ID
		req := &models.ServiceCreateInstallActionWorkflowRunRequest{
			ActionWorkflowConfigID: &configID,
			RunEnvVars:             envVars,
		}

		err := m.api.CreateInstallActionWorkflowRun(m.ctx, m.installID, req)
		if err != nil {
			return executeFormSubmittedMsg{err: err}
		}

		return executeFormSubmittedMsg{}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchInstallActionWorkflowCmd,
		m.fetchLatestConfigCmd,
		common.TickCmd(common.DefaultRefreshInterval),
		m.spinner.Tick,
	)
}

func (m *Model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *Model) resize() {
	vMarginHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 2
	threeFiffs := int(m.width * 3 / 5)
	twoFiffs := m.width - threeFiffs
	m.runsWidth = threeFiffs
	m.stepsWidth = twoFiffs

	hMargin := 2
	m.header.SetWidth(m.width - hMargin)
	m.footer.SetWidth(m.width - hMargin)

	runsListHeight := m.height - vMarginHeight
	m.runsList.SetHeight(runsListHeight)
	m.runsList.SetWidth(m.runsWidth - 3)

	m.formViewport.SetHeight(runsListHeight)
	m.formViewport.SetWidth(m.runsWidth - 3)

	vpWidth := m.width - m.runsWidth - 2
	vpHeight := m.height - vMarginHeight
	m.actionConfig.SetHeight(vpHeight)
	m.actionConfig.SetWidth(vpWidth)

	m.populateActionConfigView(true)
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.resize()
}

func (m *Model) toggleFocus() {
	if m.focus == RunsFocusArea {
		m.focus = StepsFocusArea
	} else {
		m.focus = RunsFocusArea
	}
}

func (m *Model) handleNav(msg tea.KeyPressMsg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == StepsFocusArea {
		m.actionConfig, cmd = m.actionConfig.Update(msg)
	} else {
		m.runsList, cmd = m.runsList.Update(msg)
	}
	return m, cmd
}

func (m *Model) setFormError(err error) {
	m.formError = err
	m.updateFormViewportContent()
}

func (m *Model) resetForm() {
	m.formError = nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case executeFormSubmittedMsg:
		m.formSubmitting = false
		if msg.err != nil {
			m.setLogMessage(fmt.Sprintf("Error executing action: %s", msg.err), "error")
			m.setFormError(msg.err)
		} else {
			m.setLogMessage("Action executed successfully!", "success")
			m.viewMode = RunsView
			m.keys.updateNavigationKeys(RunsView)
			return m, tea.Batch(
				m.fetchInstallActionWorkflowCmd,
				m.fetchLatestConfigCmd,
			)
		}
		return m, nil

	case common.TickMsg:
		return m, tea.Batch(
			m.fetchInstallActionWorkflowCmd,
			m.fetchLatestConfigCmd,
			common.TickCmd(common.DefaultRefreshInterval),
		)

	case installActionWorkflowFetchedMsg:
		m.handleInstallActionWorkflowFetched(msg)
	case latestConfigFetchedMsg:
		m.handleLatestConfigFetched(msg)

	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		if m.viewMode == ExecuteView {
			switch {
			case key.Matches(msg, m.keys.Quit):
				m.setQuitting()
				return m, tea.Quit
			case key.Matches(msg, m.keys.Esc):
				m.viewMode = RunsView
				m.keys.updateNavigationKeys(RunsView)
				m.setLogMessage("", "")
				m.resetForm()
				return m, nil
			case key.Matches(msg, m.keys.Tab):
				if len(m.formInputs) > 0 {
					m.nextFormInput()
					m.updateFormViewportContent()
				}
				return m, nil
			case msg.String() == "shift+tab":
				if len(m.formInputs) > 0 {
					m.prevFormInput()
					m.updateFormViewportContent()
				}
				return m, nil
			case msg.String() == "enter":
				if !m.formSubmitting {
					m.formSubmitting = true
					m.setLogMessage("Executing action...", "info")
					return m, m.submitExecuteForm()
				}
				return m, nil
			case key.Matches(msg, m.keys.Up):
				m.formViewport, cmd = m.formViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case key.Matches(msg, m.keys.Down):
				m.formViewport, cmd = m.formViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case key.Matches(msg, m.keys.PageDown):
				m.formViewport, cmd = m.formViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case key.Matches(msg, m.keys.PageUp):
				m.formViewport, cmd = m.formViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			default:
				if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
					m.formInputs[m.formFocusIndex], cmd = m.formInputs[m.formFocusIndex].Update(msg)
					cmds = append(cmds, cmd)
					m.updateFormViewportContent()
				}
			}
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.setQuitting()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Esc):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			_, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			_, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Left):
			m.toggleFocus()
		case key.Matches(msg, m.keys.Right):
			m.toggleFocus()

		case key.Matches(msg, m.keys.PageDown):
			m.actionConfig, cmd = m.actionConfig.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keys.PageUp):
			m.actionConfig, cmd = m.actionConfig.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Slash):
			m.runsList.SetShowFilter(!m.runsList.ShowFilter())
			m.runsList.Update(msg)

		case key.Matches(msg, m.keys.Enter):
			selectedItem := m.runsList.SelectedItem()
			if run, ok := selectedItem.(listRun); ok {
				return m, func() tea.Msg {
					return ac.SwitchToRunViewMsg{RunID: run.run.ID}
				}
			}

		case key.Matches(msg, m.keys.Tab):
			m.toggleFocus()

		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()

		case key.Matches(msg, m.keys.Copy):
			m.copyActionWorkflowID()

		case key.Matches(msg, m.keys.Execute):
			if !m.keys.Execute.Enabled() {
				return m, nil
			}
			if m.latestConfig != nil {
				m.initializeExecuteForm()
				m.viewMode = ExecuteView
				m.keys.updateNavigationKeys(ExecuteView)
				m.setLogMessage("Fill in the form and press Enter to execute", "info")
			}
			return m, nil

		case key.Matches(msg, m.keys.Slash):
			m.runsList.Update(msg)

		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		if m.viewMode == ExecuteView {
			m.formViewport, cmd = m.formViewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
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
	if m.installActionWorkflow == nil {
		if m.error != nil {
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.actionConfig.Height(),
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(m.error.Error()),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.actionConfig.Height(), Padding: 1, Content: "  Loading  ", Level: "info"})
		}

	} else {
		leftPanel := ""
		if m.viewMode == ExecuteView {
			leftPanel = appStyleFocus.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.renderExecuteForm())
		} else {
			if m.focus == "runs" {
				leftPanel = appStyleFocus.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.runsList.View())
			} else {
				leftPanel = appStyleBlur.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.runsList.View())
			}
		}

		stepsDetail := ""
		if m.focus == "steps" {
			stepsDetail = appStyleFocus.Render(m.actionConfig.View())
		} else {
			stepsDetail = appStyleBlur.Render(m.actionConfig.View())
		}
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftPanel,
			stepsDetail,
		)
	}
	footer := m.footerView()
	s := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	return s
}
