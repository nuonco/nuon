/*

An alt-screen TUI for viewing a workflow in detail and approving steps.

*/

package workflow

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
)

const minRequiredWidth int = 100

type model struct {
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	installID  string
	workflowID string

	// data
	workflow                     *models.AppWorkflow
	steps                        [][]*models.AppWorkflowStep // standlone so we can sort them, nested so we can group them
	selectedIndex                int                         // used to set selectedStep on data refresh (smells, use map or something better)
	selectedStep                 *models.AppWorkflowStep
	selectedStepApprovalResponse *models.AppWorkflowStepApprovalResponse
	// conditional
	stack        *models.AppInstallStack
	stackLoading bool

	// display only elements
	logMessage string

	// ui components
	// 1. layout
	header     viewport.Model
	stepsList  list.Model
	stepDetail viewport.Model
	footer     viewport.Model

	// 2. ui
	// for the header
	progress    progress.Model
	searchInput textinput.Model
	spinner     spinner.Model

	// approval confirmations
	stepApprovalConf        bool
	workflowApprovalConf    bool
	workflowCancelationConf bool
	showJson                bool

	// for the footer
	help help.Model

	// keys
	keys keyMap

	// other
	error    error
	quitting bool
	loading  bool
}

func initialStepsList() list.Model {
	stepsList := list.New([]list.Item{}, list.NewDefaultDelegate(), minRequiredWidth, 0)
	stepsList.SetShowPagination(false)
	stepsList.SetShowStatusBar(false)
	stepsList.SetShowHelp(false)
	stepsList.SetShowTitle(false)
	return stepsList
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	workflowID string,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // .Padding(0, 0, 0, 1)
	stepsList := initialStepsList()
	progress := progress.New()

	m := model{
		ctx:        ctx,
		cfg:        cfg,
		api:        api,
		installID:  installID,
		workflowID: workflowID,

		header:     viewport.New(minRequiredWidth, 2),
		stepsList:  stepsList,
		stepDetail: viewport.New(minRequiredWidth, 30),
		footer:     viewport.New(minRequiredWidth, 4),

		help:     help.New(),
		spinner:  s,
		progress: progress,

		keys: keys,
	}
	m.stepDetail.SetContent("Loading")

	return m
}

func (m *model) toggleShowJson() {
	m.showJson = !m.showJson
	m.populateStepDetailView(false)
}

func (m *model) setLogMessage(message string) {
	// for use from within m.Update
	m.logMessage = message
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick, m.spinner.Tick)
}

func (m *model) resetSelected() {
	// reset state
	m.stepApprovalConf = false
	m.workflowApprovalConf = false
	m.selectedStep = nil
	m.selectedIndex = -1
	m.showJson = false

	// toggle detail-specific key help
	m.keys.Esc.SetHelp("esc", "quit")
	m.keys.ToggleJson.SetEnabled(false)
	m.keys.OpenQuickLink.SetEnabled(false)
	m.keys.OpenTemplateLink.SetEnabled(false)

	// populate step detail view
	m.populateStepDetailView(true)
}

func (m *model) setSelected() {
	// reset any and all approval modals
	m.stepApprovalConf = false
	m.workflowApprovalConf = false
	m.showJson = false
	// grab the item from the list using the cursor
	items := m.stepsList.Items()
	if len(items) == 0 {
		return
	}
	m.selectedIndex = m.stepsList.Index()

	item := items[m.stepsList.Index()]
	// coerce to our type so we can use the niecities to grab the step details
	m.selectedStep = item.(listStep).Step()
	if m.stepIsApprovable() {
		m.keys.ApproveStep.SetEnabled(true)
	} else {
		m.setLogMessage(fmt.Sprintf("[%d:%02d] id:%s step is not approvable", m.stepsList.Cursor(), m.stepsList.Index(), m.selectedStep.ID))
		m.keys.ApproveStep.SetEnabled(false)
	}
	m.keys.Esc.SetHelp("esc", "back")
	m.keys.ToggleJson.SetEnabled(true) // enable the json toggle
	m.populateStepDetailView(true)

	// enable actions for install stack
	if m.selectedStep.StepTargetType == "install_stack_versions" {
		m.keys.OpenQuickLink.SetEnabled(true)
		m.keys.OpenTemplateLink.SetEnabled(true)
	}
}

func (m *model) setQuitting() {
	m.logMessage = "quitting ..."
	m.quitting = true
}

func (m *model) enableSearch() {
	m.searchInput.Focus()
}

func (m *model) setApprovalConfirmation() {
	m.stepApprovalConf = true
	m.loading = true
	m.setLogMessage("awaiting confirmation")
	m.populateStepDetailView(true)
}

func (m *model) resetApprovalConf() {
	m.stepApprovalConf = false
	m.loading = true
	m.setLogMessage("no confirmation received")
	m.populateStepDetailView(true)
}

func (m *model) setWorkflowCancelationConf() {
	m.loading = true
	m.setLogMessage("awaiting confirmation")
	m.workflowCancelationConf = true
	m.populateStepDetailView(true)
}

func (m *model) resetWorkflowCancelationConf() {
	m.workflowCancelationConf = false
	m.setLogMessage("no cancellation confirmation received")
	m.loading = false
	m.populateStepDetailView(true)
}

func (m *model) setWorkflowApprovalConf() {
	m.setLogMessage("awaiting confirmation")
	m.workflowApprovalConf = true
	m.populateStepDetailView(true)
}

func (m *model) resetWorkflowApprovalConf() {
	m.setLogMessage("no approval confirmation received")
	m.workflowApprovalConf = false
	m.populateStepDetailView(true)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	// when the window resizes, we need to set the width of our components
	// vertical margin height is the height of the header + the height of the footer
	vMarginHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 2
	// horizonal margin is just 2 because of the padding of 1
	hMargin := 2
	m.header.Width = msg.Width - hMargin
	m.progress.Width = msg.Width / 3
	m.footer.Width = msg.Width - hMargin

	// resize the list
	stepsListHeight := msg.Height - vMarginHeight
	m.stepsList.SetHeight(stepsListHeight)
	m.stepsList.SetWidth((msg.Width - hMargin) / 3)

	// make the detail viewport
	vpWidth := msg.Width - m.stepsList.Width() // no hmargin subtracted
	vpHeight := msg.Height - vMarginHeight
	m.stepDetail.Height = vpHeight
	m.stepDetail.Width = vpWidth

	m.populateStepDetailView(true)
}

// handle up and down
func (m *model) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.selectedStep != nil {
		m.stepDetail, cmd = m.stepDetail.Update(msg)
	} else {
		m.stepsList, cmd = m.stepsList.Update(msg)
	}
	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// var cmds []tea.Cmd
	switch msg := msg.(type) {

	// handle tick: fetch data
	case tickMsg:
		m.fetchWorkflow()
		return m, tick

	// handle re-size
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, cmd

	// handle keystrokes
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit): // "ctrl+c", "q"
			m.setQuitting()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Esc): // "esc": we overload this one a bit
			if m.stepApprovalConf {
				m.resetApprovalConf()
			} else if m.workflowCancelationConf {
				m.resetWorkflowCancelationConf()
			} else if m.workflowApprovalConf {
				m.resetWorkflowApprovalConf()
			} else if m.selectedStep != nil {
				m.resetSelected()
			} else {
				return m, tea.Quit
			}

		// actions: for a step
		case key.Matches(msg, m.keys.ToggleJson):
			m.toggleShowJson()
			return m, cmd
		case key.Matches(msg, m.keys.OpenQuickLink):
			m.openQuickLink()
		case key.Matches(msg, m.keys.OpenTemplateLink):
			m.openTemplateLink()

		// nav
		case key.Matches(msg, m.keys.Up):
			m, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			m, cmd := m.handleNav(msg)
			return m, cmd
		// these are really only for the step detail viewport
		case key.Matches(msg, m.keys.PageDown):
			m.stepDetail, cmd = m.stepDetail.Update(msg)
			// m, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.PageUp):
			m.stepDetail, cmd = m.stepDetail.Update(msg)
			// m, cmd := m.handleNav(msg)
			return m, cmd

		// additional navigation for when the step details is open/populated
		// to allow for navigation
		case key.Matches(msg, m.keys.Left):
			m.stepsList.CursorUp()
		case key.Matches(msg, m.keys.Right):
			m.stepsList.CursorDown()

		case key.Matches(msg, m.keys.Slash):
			m.stepsList.SetShowFilter(!m.stepsList.ShowFilter())
			m.stepsList.Update(msg)

		// selection
		case key.Matches(msg, m.keys.Enter):
			m.setSelected()
			m.stepsList.Update(msg)

			// data actions
		case key.Matches(msg, m.keys.ApproveStep):
			if m.stepApprovalConf {
				m.approveWorkflowStep()
			} else {
				m.setApprovalConfirmation()
			}
		case key.Matches(msg, m.keys.CancelWorkflow):
			if !m.workflowCancelationConf {
				m.setWorkflowCancelationConf()
			} else if m.workflowCancelationConf {
				m.cancelWorkflow()
			}
		case key.Matches(msg, m.keys.ApproveAll):
			if m.workflowApprovalConf {
				m.approveAll()
			} else {
				m.setWorkflowApprovalConf()
			}

		// search
		case key.Matches(msg, m.keys.Slash):
			m.enableSearch()
			m.stepsList.Update(msg)
		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()

		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "quitting " + m.spinner.View()
	}
	if m.header.Width == 0 {
		return ""

	} else if m.header.Width < minRequiredWidth {
		// TODO: make this message full screen
		return "This screen is too small, please increase the width.\n"
	}
	// this is the actual bulk of the wekr
	header := m.headerView()
	content := ""
	if m.workflow == nil { // initial load hasn't taken place
		// content = lipgloss.NewStyle().
		// 	Width(m.header.Width).
		// 	Height(m.stepDetail.Height).
		// 	Render("Loading")
		content = common.FullPageDialog(m.header.Width, m.stepDetail.Height, 1, "  Loading  ")

	} else {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.stepsList.View(),
			appStyle.Render(m.stepDetail.View()),
		)
	}
	footer := m.footerView()
	s := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	return s
}

func App(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	workflow_id string,
) {
	// initialize the model
	m := initialModel(ctx, cfg, api, install_id, workflow_id)
	m.fetchWorkflow()

	// initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
