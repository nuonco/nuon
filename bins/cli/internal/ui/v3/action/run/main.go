package run

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/action/run/steps"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"

	tea "charm.land/bubbletea/v2"
)

const (
	minRequiredWidth  int = 100
	minRequiredHeight int = 20
)

type Model struct {
	log *common.Logger
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	installID        string
	actionWorkflowID string
	runID            string

	width         int
	height        int
	stepsWidth    int
	stepsHeight   int
	sidebarWidth  int
	sidebarHeight int

	run *models.AppInstallActionWorkflowRun

	loading bool
	error   error

	focusedComponent string

	spinner   spinner.Model
	header    viewport.Model
	stepsView steps.Model
	sidebar   viewport.Model
	footer    viewport.Model

	status common.StatusBarRequest

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

		header:  viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(4)),
		footer:  viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(4)),
		sidebar: viewport.New(viewport.WithWidth(int(minRequiredWidth/3)), viewport.WithHeight(10)),
		spinner: s,
		status:  common.StatusBarRequest{Message: ""},

		loading: true,

		focusedComponent: "steps",

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
		common.TickCmd(common.DefaultRefreshInterval),
		m.spinner.Tick,
		m.stepsView.Init(),
	)
}

func (m *Model) setLogMessage(message string, level string) {
	m.status.Message = message
	m.status.Level = level
}

func (m *Model) setQuitting() {
	m.setLogMessage("quitting ...", "warning")
	m.quitting = true
}

func (m *Model) reflow() {
	vMarginHeight := lipgloss.Height(m.header.View()) + lipgloss.Height(m.footer.View()) + 6

	m.stepsWidth = int((m.width/3)*2) - 4
	m.stepsHeight = m.height - vMarginHeight
	m.stepsView.SetSize(m.stepsWidth, m.stepsHeight)

	m.sidebar.SetWidth((m.width - m.stepsWidth) - 4)
	m.sidebar.SetHeight(m.height - vMarginHeight)

	hMargin := 2
	m.header.SetWidth(m.width - hMargin)
	m.footer.SetWidth(m.width - hMargin)
	m.help.SetWidth(m.width - hMargin)
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

	case common.TickMsg:
		return m, tea.Batch(
			m.fetchInstallActionWorkflowRunCmd,
			common.TickCmd(common.DefaultRefreshInterval),
		)

	case installActionWorkflowRunFetchedMsg:
		m.handleInstallActionWorkflowRunFetched(msg)

	case tea.WindowSizeMsg:
		m.handleResize(msg)
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
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
			if m.focusedComponent == "steps" {
				m.focusedComponent = "sidebar"
			} else {
				m.focusedComponent = "steps"
			}
			m.setContent()
		default:
			if m.focusedComponent == "sidebar" {
				switch msg.String() {
				case "up", "k", "down", "j":
					m.sidebar, cmd = m.sidebar.Update(msg)
					cmds = append(cmds, cmd)
				}
			} else {
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

	sections := []string{}
	sections = append(sections, appStyle.Render(m.header.View()))

	content := ""
	if m.run == nil {
		if m.error != nil {
			content = common.FullPageDialog(common.FullPageDialogRequest{
				Width:   m.width,
				Height:  m.stepsHeight,
				Padding: 1,
				Content: lipgloss.NewStyle().Width(int(m.width/8) * 5).Padding(1).Render(m.error.Error()),
				Level:   "error",
			})
		} else {
			content = common.FullPageDialog(common.FullPageDialogRequest{Width: m.width, Height: m.stepsHeight, Padding: 1, Content: "  Loading  ", Level: "info"})
		}
	} else {
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
