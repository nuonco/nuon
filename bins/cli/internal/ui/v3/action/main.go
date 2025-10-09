/*

An alt-screen TUI for viewing action workflows and their runs.

*/

package action

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

type model struct {
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
	focus        string // one of "runs" or "steps"

	// 2. ui
	spinner spinner.Model

	// 3. for the footer
	status common.StatusBarRequest

	// for the footer
	help help.Model

	// keys
	keys keyMap

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

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)
	runsList := initialRunsList()

	m := model{
		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		installID:        installID,
		actionWorkflowID: actionWorkflowID,

		header:       viewport.New(minRequiredWidth, 2),
		runsList:     runsList,
		actionConfig: viewport.New(minRequiredWidth, 30),
		footer:       viewport.New(minRequiredWidth, 4),
		focus:        "runs",

		help:    help.New(),
		spinner: s,
		status:  common.StatusBarRequest{Message: ""},

		keys: keys,
	}
	m.actionConfig.SetContent("Loading")

	return m
}

func (m *model) setLogMessage(message string, level string) {
	// for use from within m.Update
	m.status.Message = message
	m.status.Level = level
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchInstallActionWorkflowCmd,
		m.fetchLatestConfigCmd,
		tick,
		m.spinner.Tick,
	)
}

func (m *model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *model) resize() {
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

	// make the steps detail viewport
	vpWidth := m.width - (m.runsWidth + 2) - 2 // actual width plus margin
	vpHeight := m.height - vMarginHeight
	m.actionConfig.Height = vpHeight
	m.actionConfig.Width = vpWidth

	// NOTE: called here to ensure proportions
	m.populateActionConfigView(true)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	// when the window resizes, store the dimensions of the window
	m.width = msg.Width
	m.height = msg.Height
	// then we call resize
	m.resize()
}

func (m *model) toggleFocus() {
	if m.focus == "runs" {
		m.focus = "steps"
	} else {
		m.focus = "runs"
	}
}

// handle up and down
func (m *model) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus == "steps" {
		m.actionConfig, cmd = m.actionConfig.Update(msg)
	} else {
		m.runsList, cmd = m.runsList.Update(msg)
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
			m, cmd := m.handleNav(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			m, cmd := m.handleNav(msg)
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
		// case key.Matches(msg, m.keys.Enter):
		// 	m.runsList.Update(msg)

		case key.Matches(msg, m.keys.Tab):
			m.toggleFocus()

		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()

		case key.Matches(msg, m.keys.Copy):
			m.copyActionWorkflowID()

		// search
		case key.Matches(msg, m.keys.Slash):
			m.runsList.Update(msg)

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
		runsList := ""
		if m.focus == "runs" {
			runsList = appStyleFocus.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.runsList.View())
		} else {
			runsList = appStyleBlur.Width(m.runsWidth).Padding(0, 1, 0, 0).Render(m.runsList.View())
		}
		stepsDetail := ""
		if m.focus == "steps" {
			stepsDetail = appStyleFocus.Render(m.actionConfig.View())
		} else {
			stepsDetail = appStyleBlur.Render(m.actionConfig.View())
		}
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			runsList,
			stepsDetail,
		)
	}
	footer := m.footerView()
	s := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	return s
}

func ActionWorkflowApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	action_workflow_id string,
) {
	// initialize the model
	m := initialModel(ctx, cfg, api, install_id, action_workflow_id)
	// initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
