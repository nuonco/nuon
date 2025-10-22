/*

An alt-screen TUI for viewing action workflows and their runs.

*/

package detail

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"

	ac "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/common"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

const (
	minRequiredWidth    int           = 100
	minRequiredHeight   int           = 20
	dataRefreshInterval time.Duration = time.Second * 5
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
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	installID        string
	actionWorkflowID string

	width      int
	height     int
	runsWidth  int // left section width
	stepsWidth int // right section width

	// data
	installActionWorkflow *models.AppInstallActionWorkflow // contains action workflow + runs
	latestConfig          *models.AppActionWorkflowConfig  // contains latest steps

	// loading states
	workflowLoading bool
	configLoading   bool

	// ui components
	// 1. layout
	header       viewport.Model
	runsList     list.Model
	actionConfig viewport.Model
	footer       viewport.Model
	focus        FocusArea // one of "runs" or "steps"

	// 2. ui
	spinner spinner.Model

	// 3. for the footer
	status common.StatusBarRequest

	// for the footer
	help help.Model

	// keys
	keys keyMap

	// execute form state
	viewMode       ViewMode // "runs" or "execute"
	formInputs     []textinput.Model
	formFocusIndex int
	formMappings   []executeInputMapping
	formSubmitting bool
	formError      error
	formViewport   viewport.Model

	// other
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

		header:       viewport.New(minRequiredWidth, 2),
		runsList:     runsList,
		actionConfig: viewport.New(minRequiredWidth, 30),
		footer:       viewport.New(minRequiredWidth, 4),
		focus:        RunsFocusArea,

		help:    help.New(),
		spinner: s,
		status:  common.StatusBarRequest{Message: ""},

		keys:         keys,
		viewMode:     RunsView,
		formViewport: viewport.New(minRequiredWidth, 30),
	}
	m.actionConfig.SetContent("Loading")

	return m
}

func (m *Model) setLogMessage(message string, level string) {
	// for use from within m.Update
	m.status.Message = message
	m.status.Level = level
}

func (m *Model) initializeExecuteForm() {
	m.formInputs = make([]textinput.Model, 0)
	m.formMappings = make([]executeInputMapping, 0)

	// Collect all unique env vars from all steps
	envVarsMap := make(map[string]string) // map[varName]defaultValue
	if m.latestConfig != nil && m.latestConfig.Steps != nil {
		for _, step := range m.latestConfig.Steps {
			if step == nil || step.EnvVars == nil {
				continue
			}

			for name, value := range step.EnvVars {
				// Only add if not already present (first step wins for default value)
				if _, exists := envVarsMap[name]; !exists {
					envVarsMap[name] = value
				}
			}
		}
	}

	// Sort env var names for consistent ordering
	var sortedNames []string
	for name := range envVarsMap {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Create inputs from the collected env vars in sorted order
	for _, name := range sortedNames {
		value := envVarsMap[name]

		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("Enter %s", name)
		ti.CharLimit = 500
		ti.Width = 50
		ti.Prompt = ""

		// Set default value
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

	// Focus the first input
	if len(m.formInputs) > 0 {
		m.formInputs[0].Focus()
		m.formFocusIndex = 0
	}

	// Update the form viewport content
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
		// Build env vars map
		envVars := make(map[string]string)
		for i, mapping := range m.formMappings {
			value := strings.TrimSpace(m.formInputs[i].Value())
			// Use the value if provided, otherwise use the default
			if value != "" {
				envVars[mapping.name] = value
			} else if mapping.value != "" {
				envVars[mapping.name] = mapping.value
			}
		}

		// Create the request
		req := &models.ServiceCreateInstallActionWorkflowRunRequest{
			ActionWorkflowConfigID: m.latestConfig.ID,
			RunEnvVars:             envVars,
		}

		// Call the API
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
		tick,
		m.spinner.Tick,
	)
}

func (m *Model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *Model) resize() {
	// vertical margin height is the height of the header + the height of the footer
	vMarginHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 2
	// runs take 2/3, steps take 1/3
	threeFiffs := int(m.width * 3 / 5)
	twoFiffs := m.width - threeFiffs
	m.runsWidth = threeFiffs
	m.stepsWidth = twoFiffs

	// horizonal margin is just 2 because of the padding of 1
	hMargin := 2
	m.header.Width = m.width - hMargin
	m.footer.Width = m.width - hMargin

	// resize the runs list
	runsListHeight := m.height - vMarginHeight
	m.runsList.SetHeight(runsListHeight)
	m.runsList.SetWidth(m.runsWidth - 1) // minus one because of the padding we render the list with

	// resize the form viewport (same height as runs list)
	m.formViewport.Height = runsListHeight
	m.formViewport.Width = m.runsWidth - 4 // account for padding

	// make the steps detail viewport
	vpWidth := m.width - (m.runsWidth + 2) - 2 // actual width plus margin
	vpHeight := m.height - vMarginHeight
	m.actionConfig.Height = vpHeight
	m.actionConfig.Width = vpWidth

	// NOTE: called here to ensure proportions
	m.populateActionConfigView(true)
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	// when the window resizes, store the dimensions of the window
	m.width = msg.Width
	m.height = msg.Height
	// then we call resize
	m.resize()
}

func (m *Model) toggleFocus() {
	if m.focus == RunsFocusArea {
		m.focus = StepsFocusArea
	} else {
		m.focus = RunsFocusArea
	}
}

// handle up and down
func (m *Model) handleNav(msg tea.KeyMsg) (*Model, tea.Cmd) {
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
			// Switch back to runs view and refresh data
			m.viewMode = RunsView
			m.keys.updateNavigationKeys(RunsView)
			return m, tea.Batch(
				m.fetchInstallActionWorkflowCmd,
				m.fetchLatestConfigCmd,
			)
		}
		return m, nil

	// handle tick: data refresh and ticks
	case tickMsg:
		return m, tea.Batch(
			m.fetchInstallActionWorkflowCmd,
			m.fetchLatestConfigCmd,
			tea.Tick(
				dataRefreshInterval,
				func(t time.Time) tea.Msg {
					return tickMsg(t)
				}),
		)

	case installActionWorkflowFetchedMsg:
		m.handleInstallActionWorkflowFetched(msg)
	case latestConfigFetchedMsg:
		m.handleLatestConfigFetched(msg)

	// handle re-size
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	// handle keystrokes
	case tea.KeyMsg:
		// Handle execute mode keys
		if m.viewMode == ExecuteView {
			switch {
			case key.Matches(msg, m.keys.Quit): // "ctrl+c", "q"
				m.setQuitting()
				return m, tea.Quit
			case key.Matches(msg, m.keys.Esc): // "esc": go back to runs view
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
				// Up arrow scrolls the viewport
				m.formViewport, cmd = m.formViewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case key.Matches(msg, m.keys.Down):
				// Down arrow scrolls the viewport
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
				// Handle text input
				if m.formFocusIndex >= 0 && m.formFocusIndex < len(m.formInputs) {
					m.formInputs[m.formFocusIndex], cmd = m.formInputs[m.formFocusIndex].Update(msg)
					cmds = append(cmds, cmd)
					m.updateFormViewportContent()
				}
			}
			return m, tea.Batch(cmds...)
		}

		// Handle runs mode keys
		switch {
		case key.Matches(msg, m.keys.Quit): // "ctrl+c", "q"
			m.setQuitting()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Esc): // "esc": we overload this one a bit
			return m, tea.Quit

		// nav
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

		// these are really only for the steps detail viewport
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

		// selection
		case key.Matches(msg, m.keys.Enter):
			selectedItem := m.runsList.SelectedItem()
			if run, ok := selectedItem.(listRun); ok {
				// Return the message to parent
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
			// Only allow execution if we have a config and it's not disabled
			if !m.keys.Execute.Enabled() {
				return m, nil
			}
			// Switch to execute mode
			if m.latestConfig != nil {
				m.initializeExecuteForm()
				m.viewMode = ExecuteView
				m.keys.updateNavigationKeys(ExecuteView)
				m.setLogMessage("Fill in the form and press Enter to execute", "info")
			}
			return m, nil

		// search
		case key.Matches(msg, m.keys.Slash):
			m.runsList.Update(msg)

		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Handle mouse events for viewport scrolling in execute mode
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

	// this is the actual bulk of the work
	header := m.headerView()
	content := ""
	if m.installActionWorkflow == nil { // initial load hasn't taken place
		if m.error != nil { // likely a 404 but worth refining later
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.actionConfig.Height,
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(fmt.Sprintf("%s", m.error.Error())),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.actionConfig.Height, Padding: 1, Content: "  Loading  ", Level: "info"})
		}

	} else {
		leftPanel := ""
		if m.viewMode == ExecuteView {
			// Render execute form in left panel
			leftPanel = appStyleFocus.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.renderExecuteForm())
		} else {
			// Render runs list in left panel
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
