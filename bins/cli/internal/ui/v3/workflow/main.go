/*

An alt-screen TUI for viewing a workflow in detail and approving steps.

*/

package workflow

import (
	"context"
	"fmt"
	"os"
	"time"

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
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

const (
	minRequiredWidth    int           = 100
	minRequiredHeight   int           = 20
	dataRefreshInterval time.Duration = time.Second * 5
)

type approvalContents struct {
	raw      interface{}
	contents map[string]interface{}
	loading  bool
	error    error
	// target attrs for decompression
	terraformContent map[string]any
	helmDiff         map[string]any
}

type model struct {
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	installID  string
	workflowID string

	width     int
	height    int
	listWidth int

	// data
	workflow                     *models.AppWorkflow
	steps                        [][]*models.AppWorkflowStep // standlone so we can sort them, nested so we can group them
	selectedIndex                int                         // used to set selectedStep on data refresh (smells, use map or something better)
	selectedStep                 *models.AppWorkflowStep
	selectedStepApprovalResponse *models.ServiceCreateWorkflowStepApprovalResponseResponse

	// conditional
	stack        *models.AppInstallStack
	stackLoading bool

	approvalContents approvalContents

	// ui components
	// 1. layout
	header     viewport.Model
	stepsList  list.Model
	stepDetail viewport.Model
	footer     viewport.Model
	focus      string // one of "list" or "detail"

	// 2. ui
	// for the header
	progress    progress.Model
	searchInput textinput.Model
	spinner     spinner.Model

	// 3. for the footer
	status common.StatusBarRequest

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
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor) // .Padding(0, 0, 0, 1)
	stepsList := initialStepsList()
	progress := progress.New()
	approvalContents := approvalContents{error: nil, loading: false, raw: []int64{}}

	m := model{
		ctx:        ctx,
		cfg:        cfg,
		api:        api,
		installID:  installID,
		workflowID: workflowID,

		// data
		approvalContents: approvalContents,

		header:     viewport.New(minRequiredWidth, 2),
		stepsList:  stepsList,
		stepDetail: viewport.New(minRequiredWidth, 30),
		footer:     viewport.New(minRequiredWidth, 4),
		focus:      "list",

		help:     help.New(),
		spinner:  s,
		progress: progress,
		status:   common.StatusBarRequest{Message: ""},

		keys: keys,
	}
	m.stepDetail.SetContent("Loading")

	return m
}

func (m *model) toggleShowJson() {
	m.showJson = !m.showJson
	m.populateStepDetailView(false)
}

func (m *model) setLogMessage(message string, level string) {
	// for use from within m.Update
	m.status.Message = message
	m.status.Level = level
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchWorkflowCmd,
		m.fetchStackCmd,
		tick,
		m.spinner.Tick,
	)
}

func (m *model) resetSelected() {
	// reset state
	m.stepApprovalConf = false
	m.workflowApprovalConf = false
	m.selectedStep = nil
	m.selectedIndex = -1
	m.showJson = false
	m.approvalContents = approvalContents{loading: false, error: nil}

	// toggle detail-specific key help
	m.keys.Esc.SetHelp("esc", "quit")
	m.keys.ToggleJson.SetEnabled(false)
	m.keys.OpenQuickLink.SetEnabled(false)
	m.keys.OpenTemplateLink.SetEnabled(false)

	// populate step detail view
	m.populateStepDetailView(true)
	m.focus = "list"
}

func (m *model) setSelected() []tea.Cmd {
	cmds := []tea.Cmd{}
	// reset any and all approval modals
	m.stepApprovalConf = false
	m.workflowApprovalConf = false
	m.showJson = false
	// grab the item from the list using the cursor
	items := m.stepsList.Items()
	if len(items) == 0 {
		return cmds
	}
	m.selectedIndex = m.stepsList.Index()

	item := items[m.stepsList.Index()]
	// coerce to our type so we can use the niecities to grab the step details
	m.selectedStep = item.(listStep).Step()
	if m.stepIsApprovable() {
		m.keys.ApproveStep.SetEnabled(true)
	} else {
		m.setLogMessage(
			fmt.Sprintf(
				"[%02d] id:%s step is not approvable",
				m.stepsList.Index(),
				m.selectedStep.ID,
			),
			"info",
		)
		m.keys.ApproveStep.SetEnabled(false)
	}
	m.keys.Esc.SetHelp("esc", "back")
	m.keys.ToggleJson.SetEnabled(true) // enable the json toggle
	m.populateStepDetailView(true)

	// enable actions for install stack
	switch m.selectedStep.StepTargetType {
	case "install_stack_versions":
		m.keys.OpenQuickLink.SetEnabled(true)
		m.keys.OpenTemplateLink.SetEnabled(true)
	case "install_deploys":
		if m.selectedStep.Approval != nil {
			m.setLogMessage(
				fmt.Sprintf(
					"[%02d] id:%s fetching approval contents",
					m.stepsList.Index(),
					m.selectedStep.ID,
				),
				"info",
			)
			m.approvalContents.loading = true
			cmds = append(cmds, m.getWorkflowStepApprovalContentsCmd)
		}

	}

	m.focus = "detail"
	return cmds
}

func (m *model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *model) enableSearch() {
	m.searchInput.Focus()
}

func (m *model) setApprovalConfirmation() {
	m.stepApprovalConf = true
	m.loading = true
	m.setLogMessage("awaiting confirmation", "info")
	m.populateStepDetailView(true)
}

func (m *model) resetApprovalConf() {
	m.stepApprovalConf = false
	m.loading = true
	m.setLogMessage("no confirmation received", "warning")
	m.populateStepDetailView(true)
}

func (m *model) setWorkflowCancelationConf() {
	m.loading = true
	m.setLogMessage("awaiting confirmation", "info")
	m.workflowCancelationConf = true
	m.populateStepDetailView(true)
}

func (m *model) resetWorkflowCancelationConf() {
	m.workflowCancelationConf = false
	m.setLogMessage("no cancellation confirmation received", "warning")
	m.loading = false
	m.populateStepDetailView(true)
}

func (m *model) setWorkflowApprovalConf() {
	m.setLogMessage("awaiting confirmation", "info")
	m.workflowApprovalConf = true
	m.populateStepDetailView(true)
}

func (m *model) resetWorkflowApprovalConf() {
	m.setLogMessage("no approval confirmation received", "warning")
	m.workflowApprovalConf = false
	m.populateStepDetailView(true)
}

func (m *model) resize() {
	// vertical margin height is the height of the header + the height of the footer
	vMarginHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 2
	third := int(m.width / 3)
	// the list width controls the width of the style.Width we render the list with
	m.listWidth = third
	// horizonal margin is just 2 because of the padding of 1
	hMargin := 2
	m.header.Width = m.width - hMargin
	m.progress.Width = third
	m.footer.Width = m.width - hMargin

	// resize the list
	stepsListHeight := m.height - vMarginHeight
	m.stepsList.SetHeight(stepsListHeight)
	m.stepsList.SetWidth(m.listWidth - 1) // minus one because of the padding we render the list with

	// make the detail viewport
	vpWidth := m.width - (m.listWidth + 2) - 2 // actual width plus margin
	vpHeight := m.height - vMarginHeight
	m.stepDetail.Height = vpHeight
	m.stepDetail.Width = vpWidth

	// NOTE: called here to ensure proportions
	m.populateStepDetailView(true)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	// when the window resizes, store the dimensions of the window
	m.width = msg.Width
	m.height = msg.Height
	// then we call resize
	m.resize()
}

func (m *model) toggleFocus() {
	if m.focus == "list" {
		m.focus = "detail"
	} else {
		m.focus = "list"
	}
}

// handle up and down

func (m *model) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == "detail" { // m.selectedStep != nil
		m.stepDetail, cmd = m.stepDetail.Update(msg)
	} else {
		m.stepsList, cmd = m.stepsList.Update(msg)
	}
	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// handle tick: data refresh and ticks
	case tickMsg:
		return m, tea.Batch(
			m.fetchWorkflowCmd,
			m.fetchStackCmd,
			tea.Tick(
				dataRefreshInterval,
				func(t time.Time) tea.Msg {
					return tickMsg(t)
				}),
		)

	case workflowFetchedMsg:
		m.handleWorkflowFetched(msg)
	case stackFetchedMsg:
		m.handleStackFetched(msg)
	case createWorkflowStepApprovalResponseMsg:
		cmd = m.handleWorkflowStepApprovalResponseCreated(msg)
		cmds = append(cmds, cmd)
	case cancelWorkflowMsg:
		cmd = m.handleCancelWorkflow(msg)
		cmds = append(cmds, cmd)
	case approveAllMsg:
		commands := m.handleApproveAll(msg)
		cmds = append(cmds, commands...)
	case getWorkflowStepApprovalContentsMsg:
		commands := m.handleGetWorkflowStepApprovalContents(msg)
		cmds = append(cmds, commands...)

	// handle re-size
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

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
			return m, tea.Batch(cmds...)
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
		case key.Matches(msg, m.keys.Left):
			m.toggleFocus()
		case key.Matches(msg, m.keys.Right):
			m.toggleFocus()

		// these are really only for the step detail viewport
		case key.Matches(msg, m.keys.PageDown):
			m.stepDetail, cmd = m.stepDetail.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keys.PageUp):
			m.stepDetail, cmd = m.stepDetail.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Slash):
			m.stepsList.SetShowFilter(!m.stepsList.ShowFilter())
			m.stepsList.Update(msg)

		// selection
		case key.Matches(msg, m.keys.Enter):
			commands := m.setSelected()
			cmds = append(cmds, commands...)
			m.stepsList.Update(msg)

			// data actions
		case key.Matches(msg, m.keys.ApproveStep):
			if m.stepApprovalConf {
				m.createWorkflowStepApprovalResponseCmd()
			} else {
				m.setApprovalConfirmation()
			}
		case key.Matches(msg, m.keys.CancelWorkflow):
			if !m.workflowCancelationConf {
				m.setWorkflowCancelationConf()
			} else if m.workflowCancelationConf {
				cmds = append(cmds, m.cancelWorkflowCmd)
			}
		case key.Matches(msg, m.keys.ApproveAll):
			if m.workflowApprovalConf {
				cmds = append(cmds, m.approveAllCmd)
			} else {
				m.setWorkflowApprovalConf()
			}

		case key.Matches(msg, m.keys.Tab):
			m.toggleFocus()

		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()

		// search
		case key.Matches(msg, m.keys.Slash):
			m.enableSearch()
			m.stepsList.Update(msg)

		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
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
	if m.workflow == nil { // initial load hasn't taken place
		if m.error != nil { // likely a 404 but worth refining later
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.stepDetail.Height,
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(fmt.Sprintf("%s", m.error.Error())),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.stepDetail.Height, Padding: 1, Content: "  Loading  ", Level: "info"})
		}

	} else {
		stepsList := ""
		if m.focus == "list" {
			stepsList = appStyleFocus.Width(m.listWidth).Padding(0, 1, 0, 0).Render(m.stepsList.View())
		} else {
			stepsList = appStyleBlur.Width(m.listWidth).Padding(0, 1, 0, 0).Render(m.stepsList.View())
		}
		stepDetail := ""
		if m.focus == "detail" {
			stepDetail = appStyleFocus.Render(m.stepDetail.View())
		} else {
			stepDetail = appStyleBlur.Render(m.stepDetail.View())
		}
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			stepsList,
			stepDetail,
		)
	}
	footer := m.footerView()
	s := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	return s
}

func WorkflowApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	workflow_id string,
) {
	// initialize the model
	m := initialModel(ctx, cfg, api, install_id, workflow_id)
	// initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
