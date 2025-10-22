package run

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
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

func (m model) View() string {
	return m.m.View()
}

func ActionWorkflowRunApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	action_workflow_id string,
	run_id string,
) {
	// initialize the model
	app := initialModel(ctx, cfg, api, install_id, action_workflow_id, run_id)
	m := model{m: app}
	// initialize the program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		// tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
