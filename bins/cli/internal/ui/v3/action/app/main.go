package app

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"

	"github.com/nuonco/nuon-go"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	ac "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/common"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/detail"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/run"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
)

type ActionView string

const (
	Detail ActionView = "detail"
	Run    ActionView = "run"
)

type model struct {
	log *common.Logger
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	keys keyMap

	installID        string
	actionWorkflowID string

	// track window dimensions
	width  int
	height int

	view       ActionView
	detailView detail.Model
	runView    run.Model
}

func (m model) Init() tea.Cmd {
	return m.detailView.Init()
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	actionWorkflowID string,
) model {
	log, _ := common.NewLogger("install-action")
	detailView := detail.New(
		ctx,
		cfg,
		api,
		installID,
		actionWorkflowID,
	)
	m := model{
		log:              log,
		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		keys:             keys,
		installID:        installID,
		actionWorkflowID: actionWorkflowID,
		view:             Detail,
		detailView:       detailView,
	}
	return m
}

func (m *model) setView(view ActionView) {
	m.view = view
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// var cmds []tea.Cmd

	// these cases are the only cases this app handles
	// for all other cases , we just pass the msg to the active view
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Track window dimensions in parent
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit, m.keys.Esc):
			if m.view == Run { // if we are in a run
				if m.runView.ExpandedStepIndex() == -1 { // if there are no expanded/selected  steps
					m.log.Info("switching to detail view")
					// m.detailView = detail.New(
					// 	m.ctx,
					// 	m.cfg,
					// 	m.api,
					// 	m.installID,
					// 	m.actionWorkflowID,
					// )
					m.setView(Detail)
					// Send resize message to new detail view so it has correct dimensions
					resizeCmd := func() tea.Msg {
						return tea.WindowSizeMsg{Width: m.width, Height: m.height}
					}
					return m, tea.Batch(m.detailView.Init(), resizeCmd)
				} else { // if a step IS expanded
					m.runView, cmd = m.runView.Update(msg)
				}
			} else { // if we are in the detail view, go ahead and quit
				return m, tea.Quit
			}
		}

	case ac.SwitchToRunViewMsg:
		m.log.Info("switching to run view", zap.String("run_id", msg.RunID))
		m.runView = run.New(m.ctx, m.cfg, m.api, m.installID, m.actionWorkflowID, msg.RunID)
		m.setView(Run)
		// Send resize message to new run view so it has correct dimensions
		resizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}
		return m, tea.Batch(m.runView.Init(), resizeCmd)
	}

	switch m.view {
	case Detail:
		m.detailView, cmd = m.detailView.Update(msg)
	case Run:
		m.runView, cmd = m.runView.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	switch m.view {
	case Detail:
		return m.detailView.View()
	case Run:
		return m.runView.View()
	default:
		return "what is you talm bout"
	}
}

func App(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	action_workflow_id string,
) {
	m := initialModel(ctx, cfg, api, install_id, action_workflow_id)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
