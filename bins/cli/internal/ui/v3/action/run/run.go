package run

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type model struct {
	m Model
}

func (m model) Init() tea.Cmd {
	cmd := m.m.Init()
	return cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.m, cmd = m.m.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.m.View())
	v.AltScreen = true
	return v
}

func ActionWorkflowRunApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	action_workflow_id string,
	run_id string,
) {
	if !cfg.Interactive {
		fmt.Fprintln(os.Stderr, "interactive terminal required; use --json flag for non-interactive output")
		os.Exit(1)
	}

	// initialize the model
	app := initialModel(ctx, cfg, api, install_id, action_workflow_id, run_id)
	m := model{m: app}
	// initialize the program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
