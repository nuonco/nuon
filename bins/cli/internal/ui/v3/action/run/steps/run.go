/*
This run.go file contains a stub used to render the step component in a standalone way.
The intention is for it to be used by the forthcoming storybook.
*/
package steps

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
)

// model wraps the Model for bubbletea
type model struct {
	m Model
}

func (m model) Init() tea.Cmd {
	return m.m.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.m, cmd = m.m.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.m.View()
}

// NOTE(fd): for the storybook or any other scenario where you want to run as a standalone bubbletea program
func Run(
	ctx context.Context,
	api nuon.Client,
	width int,
	height int,
	run *models.AppInstallActionWorkflowRun,
) {
	// Initialize the model
	app := New(ctx, api, width, height, run)
	m := model{m: app}

	// Initialize the program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running steps app: %v", err)
		os.Exit(1)
	}
}
