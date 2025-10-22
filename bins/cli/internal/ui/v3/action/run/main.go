/*

An alt-screen TUI for viewing a specific action run.

*/

package run

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/run/steps"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
	"go.uber.org/zap"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	minRequiredWidth    int           = 100
	minRequiredHeight   int           = 20
	dataRefreshInterval time.Duration = time.Second * 5
)

type Model struct {
	// internal
	log *common.Logger
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	installID        string
	actionWorkflowID string
	runID            string

	// sizing
	width         int
	height        int
	stepsWidth    int // 2/3 of available width
	stepsHeight   int
	sidebarWidth  int // 1/3 of available width
	sidebarHeight int

	// data
	run *models.AppInstallActionWorkflowRun

	// loading states
	loading bool
	error   error

	// focus state
	focusedComponent string // "steps" or "sidebar"

	// ui components
	spinner   spinner.Model
	header    viewport.Model
	stepsView steps.Model // component: steplist with logs + selection & expansion
	sidebar   viewport.Model
	footer    viewport.Model

	// for the footer
	status common.StatusBarRequest

	// other
	help     help.Model
	keys     keyMap
	quitting bool
}

func (m Model) ExpandedStepIndex() int {
	return m.stepsView.ExpandedStepIndex()
}

func New(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
	runID string,
) Model {
	log, _ := common.NewLogger("install-action-run")
	return Model{
		log:              log,
		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		installID:        installID,
		actionWorkflowID: actionWorkflowID,
		runID:            runID,
	}

}
func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
	runID string,
) Model {
	log, _ := common.NewLogger("install-action-run")
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	m := Model{
		log: log,

		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		installID:        installID,
		actionWorkflowID: actionWorkflowID,
		runID:            runID,

		header:  viewport.New(minRequiredWidth, 4),
		footer:  viewport.New(minRequiredWidth, 4),
		sidebar: viewport.New(int(minRequiredWidth/3), 10),
		spinner: s,
		status:  common.StatusBarRequest{Message: ""},

		loading: true,

		focusedComponent: "steps", // Start with steps focused

		help: help.New(),
		keys: keys,
	}

	m.stepsView = steps.New(m.ctx, m.api, m.stepsWidth, m.stepsHeight, m.run)
	log.Info("initialModel: set steps view dimensions", zap.Int("string", m.stepsWidth), zap.Int("height", m.stepsHeight))
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchInstallActionWorkflowRunCmd,
		tick,
		m.spinner.Tick,
		m.stepsView.Init(),
	)
}

func (m *Model) setLogMessage(message string, level string) {
	// for use from within m.Update
	m.status.Message = message
	m.status.Level = level
}

func (m *Model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *Model) reflow() {
	// apply sizing to children
	// vertical margin height is the height of the header + the height of the footer
	vMarginHeight := lipgloss.Height(m.header.View()) + lipgloss.Height(m.footer.View()) + 6

	// steps take full width minus padding
	// the steps should be 2/3
	m.stepsWidth = int((m.width/3)*2) - 4
	m.stepsHeight = m.height - vMarginHeight
	m.stepsView.SetSize(m.stepsWidth, m.stepsHeight)

	m.sidebar.Width = (m.width - m.stepsWidth) - 4
	m.sidebar.Height = m.height - vMarginHeight

	// horizontal margin is just 2 because of the padding of 1
	hMargin := 2
	m.header.Width = m.width - hMargin
	m.footer.Width = m.width - hMargin
	m.help.Width = m.width - hMargin
}

func (m *Model) setContent() {
	m.reflow()
	m.setHeaderContent()
	m.setSidebarContent()
	m.setFooterContent()
	m.reflow()
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.log.Info("handling resize event", zap.Int("msg.width", msg.Width), zap.Int("msg.height", msg.Height))
	m.width = msg.Width
	m.height = msg.Height
	m.reflow()
	// Update viewport content after resize
	m.log.Info("handled resize event", zap.Int("m.width", m.width), zap.Int("m.height", m.height))
}

func (m *Model) toggleHelp() {
	m.log.Info("toggling help", zap.Bool("ShowAll", !m.help.ShowAll))
	m.help.ShowAll = !m.help.ShowAll
	m.setFooterContent()
	m.reflow()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// handle tick: data refresh
	case tickMsg:
		return m, tea.Batch(
			m.fetchInstallActionWorkflowRunCmd,
			tea.Tick(
				dataRefreshInterval,
				func(t time.Time) tea.Msg {
					return tickMsg(t)
				}),
		)

	case installActionWorkflowRunFetchedMsg:
		m.handleInstallActionWorkflowRunFetched(msg)

	// handle re-size
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	// handle keystrokes
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.setQuitting()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Esc):
			m.setQuitting()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.toggleHelp()
		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()
		case key.Matches(msg, m.keys.Tab):
			// Toggle focus between steps and sidebar
			if m.focusedComponent == "steps" {
				m.focusedComponent = "sidebar"
			} else {
				m.focusedComponent = "steps"
			}
			m.setContent() // Update styles based on new focus
		default:
			// Route input based on focused component
			if m.focusedComponent == "sidebar" {
				// Handle sidebar scrolling with arrow keys
				switch msg.String() {
				case "up", "k", "down", "j":
					m.sidebar, cmd = m.sidebar.Update(msg)
					cmds = append(cmds, cmd)
				}
			} else {
				// Default: pass to stepsView
				m.stepsView, cmd = m.stepsView.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		m.stepsView, cmd = m.stepsView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return "quitting " + m.spinner.View()
	}

	if m.width == 0 {
		return ""
	}

	if m.width < minRequiredWidth || m.height < minRequiredHeight {
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
	sections := []string{}
	sections = append(sections, appStyle.Render(m.header.View()))

	content := ""
	if m.run == nil { // initial load hasn't taken place
		if m.error != nil { // likely a 404 but worth refining later
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.stepsHeight,
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(fmt.Sprintf("%s", m.error.Error())),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.stepsHeight, Padding: 1, Content: "  Loading  ", Level: "info"})
		}
	} else {
		// Apply focus styles based on which component is focused
		var stepsContent, sidebarContent string
		if m.focusedComponent == "steps" {
			stepsContent = appStyleFocus.Render(m.stepsView.View())
			sidebarContent = appStyleBlur.Render(m.sidebar.View())
		} else {
			stepsContent = appStyleBlur.Render(m.stepsView.View())
			sidebarContent = appStyleFocus.Render(m.sidebar.View())
		}

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			stepsContent,
			sidebarContent,
		)
	}
	sections = append(sections, content)
	sections = append(sections, m.footer.View())
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
