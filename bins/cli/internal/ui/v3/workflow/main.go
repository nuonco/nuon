/*

An alt-screen TUI for viewing a workflow in detail and approving steps.

*/

package workflow

import (
	"context"
	"fmt"
	"os"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

const (
	minRequiredWidth  int = 100
	minRequiredHeight int = 20
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
	policyNames      map[string]string
	helmDiffExplorer helmDiffExplorerModel

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

	// approval processing state
	approvingStep bool // whether an approval request is in flight

	// auto-retry
	autoRetryAll  bool // when true, failed retryable steps are retried automatically on each poll
	retryInFlight bool // true after a retry is fired; cleared once the step leaves error state

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
		policyNames:      map[string]string{},
		helmDiffExplorer: newHelmDiffExplorerModel(minRequiredWidth - 4),

		header:     viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(2)),
		stepsList:  stepsList,
		stepDetail: viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(30)),
		footer:     viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(4)),
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

func (m model) PolicyNameByID(id string) string {
	if id == "" {
		return ""
	}
	if name, ok := m.policyNames[id]; ok && name != "" {
		return name
	}
	return id
}

func (m *model) toggleShowJson() {
	m.showJson = !m.showJson
	m.populateStepDetailView(false)
}

func (m *model) toggleAutoRetryAll() {
	m.autoRetryAll = !m.autoRetryAll
	if m.autoRetryAll {
		m.setLogMessage("auto-retry: on — will retry failed steps on each refresh", "info")
	} else {
		m.setLogMessage("auto-retry: off", "warning")
	}
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
		common.TickCmd(common.DefaultRefreshInterval),
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
	m.helmDiffExplorer.Reset()

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
	m.approvalContents = approvalContents{loading: false, error: nil}
	m.helmDiffExplorer.Reset()
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
	if m.selectedStep.StepTargetType == "install_stack_versions" {
		m.keys.OpenQuickLink.SetEnabled(true)
		m.keys.OpenTemplateLink.SetEnabled(true)
	}

	if m.stepHasPlanDiff(m.selectedStep) && m.selectedStep.Approval != nil {
		m.setLogMessage(
			fmt.Sprintf(
				"[%02d] id:%s fetching approval contents",
				m.stepsList.Index(),
				m.selectedStep.ID,
			),
			"info",
		)
		m.approvalContents = approvalContents{loading: true, error: nil}
		cmds = append(cmds, m.getWorkflowStepApprovalContentsCmd)
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
	m.header.SetWidth(m.width - hMargin)
	m.progress.SetWidth(third)
	m.footer.SetWidth(m.width - hMargin)

	// resize the list
	stepsListHeight := m.height - vMarginHeight
	m.stepsList.SetHeight(stepsListHeight)
	// Width(listWidth) is total outer width including borders (2) and padding (1 right),
	// so the list content area is listWidth - 3.
	m.stepsList.SetWidth(m.listWidth - 3)

	// make the detail viewport: total width minus list pane (listWidth) minus detail borders (2)
	vpWidth := m.width - m.listWidth - 2
	vpHeight := m.height - vMarginHeight
	m.stepDetail.SetHeight(vpHeight)
	m.stepDetail.SetWidth(vpWidth)
	m.helmDiffExplorer.SetWidth(m.stepDetail.Width() - 4)

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

func (m *model) handleNav(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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
	case common.TickMsg:
		return m, tea.Batch(
			m.fetchWorkflowCmd,
			m.fetchStackCmd,
			common.TickCmd(common.DefaultRefreshInterval),
		)

	case workflowFetchedMsg:
		cmd = m.handleWorkflowFetched(msg)
		cmds = append(cmds, cmd)

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
	case retryAllMsg:
		commands := m.handleRetryAll(msg)
		cmds = append(cmds, commands...)
	case getWorkflowStepApprovalContentsMsg:
		commands := m.handleGetWorkflowStepApprovalContents(msg)
		cmds = append(cmds, commands...)

	// handle re-size
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	// handle keystrokes
	case tea.KeyPressMsg:
		if m.handleDetailContentKey(msg) {
			m.populateStepDetailView(false)
			return m, tea.Batch(cmds...)
		}

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
			if m.focus == "detail" {
				return m, tea.Batch(cmds...)
			}

			commands := m.setSelected()
			cmds = append(cmds, commands...)
			m.stepsList.Update(msg)

			// data actions
		case key.Matches(msg, m.keys.ApproveStep):
			if m.stepApprovalConf {
				m.approvingStep = true
				// Capture values needed for the API call before returning
				cmd := m.makeApproveStepCmd()
				cmds = append(cmds, cmd)
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
				m.approvingStep = true
				cmds = append(cmds, m.approveAllCmd)
			} else {
				m.setWorkflowApprovalConf()
			}
		case key.Matches(msg, m.keys.ToggleRetryAll):
			m.toggleAutoRetryAll()

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
		// Update spinner view on list items so in-progress steps animate in sync with the header
		m.updateListSpinnerViews()
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

	// this is the actual bulk of the work
	header := m.headerView()
	content := ""
	if m.workflow == nil { // initial load hasn't taken place
		if m.error != nil { // likely a 404 but worth refining later
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.stepDetail.Height(),
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(m.error.Error()),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.stepDetail.Height(), Padding: 1, Content: "  Loading  ", Level: "info"})
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
	autoRetry bool,
) {
	if !cfg.Interactive {
		workflowPlainText(ctx, api, workflow_id)
		return
	}

	// initialize the model
	m := initialModel(ctx, cfg, api, install_id, workflow_id)
	m.autoRetryAll = autoRetry
	// initialize the program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}

// workflowPlainText fetches a workflow once and prints a plain-text summary.
func workflowPlainText(ctx context.Context, api nuon.Client, workflowID string) {
	wf, err := api.GetWorkflow(ctx, workflowID)
	if err != nil {
		fmt.Printf("Error fetching workflow: %v\n", err)
		return
	}

	status := ""
	if wf.Status != nil {
		status = string(wf.Status.Status)
	}

	fmt.Printf("Workflow: %s (%s)\n", wf.Name, wf.ID)
	fmt.Printf("Type:     %s\n", string(wf.Type))
	fmt.Printf("Status:   %s\n", status)

	if wf.Finished {
		fmt.Printf("Finished: yes\n")
	} else {
		fmt.Printf("Finished: no\n")
	}

	if len(wf.Steps) > 0 {
		fmt.Println("\nSteps:")
		for _, step := range wf.Steps {
			stepStatus := ""
			if step.Status != nil {
				stepStatus = string(step.Status.Status)
			}
			finished := "no"
			if step.Finished {
				finished = "yes"
			}
			fmt.Printf("  [%d] %-30s  %-20s  finished=%s\n", step.Idx, step.Name, stepStatus, finished)
		}
	}
}
