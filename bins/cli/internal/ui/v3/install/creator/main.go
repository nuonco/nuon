/*

An alt-screen TUI for creating installs with dynamic form generation based on app inputs.

*/

package creator

import (
	"context"
	"errors"
	"fmt"
	"os"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	maxWidth              = 80
	minRequiredWidth  int = 60
	minRequiredHeight int = 16
)

type model struct {
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	appID string

	width  int
	height int

	// data
	inputConfig *models.AppAppInputConfig
	app         *models.AppApp

	// form state
	inputs        []textinput.Model
	focusIndex    int
	regionIndex   int
	inputMappings []inputMapping

	// ui components
	viewport viewport.Model
	spinner  spinner.Model
	help     help.Model
	status   common.StatusBarRequest

	// state
	loading    bool
	submitting bool
	error      error
	success    bool
	installID  string
	quitting   bool
	keys       keyMap
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	appID string,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	vp := viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(minRequiredHeight))
	vp.YPosition = 0

	m := model{
		ctx:      ctx,
		cfg:      cfg,
		api:      api,
		appID:    appID,
		viewport: vp,
		spinner:  s,
		help:     help.New(),
		status:   common.StatusBarRequest{Message: "Loading app configuration..."},
		loading:  true,
		keys:     keys,
	}

	return m
}

func (m *model) setLogMessage(message string, level string) {
	m.status.Message = message
	m.status.Level = level
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchConfigCmd(m),
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case configFetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err
			m.setLogMessage(fmt.Sprintf("Error loading config: %s", msg.err), "error")
		} else {
			m.inputConfig = msg.inputConfig
			m.app = msg.app
			m.createFormInputs()
			m.setLogMessage("Fill in the form and press Enter to create install", "info")
		}
		return m, nil

	case installCreatedMsg:
		m.submitting = false
		if msg.err != nil {
			m.error = msg.err
			m.setLogMessage(fmt.Sprintf("Error creating install: %s", msg.err), "error")
		} else {
			m.success = true
			m.installID = msg.install.ID
			m.setLogMessage(fmt.Sprintf("Install created successfully: %s", msg.install.ID), "success")
			return m, autoExitAfterDelay()
		}
		return m, nil

	case autoExitMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		helpHeight := lipgloss.Height(m.help.View(m.keys))
		statusHeight := 3
		m.viewport.SetWidth(msg.Width - 4)
		m.viewport.SetHeight(msg.Height - helpHeight - statusHeight - 2)

		return m, nil

	case tea.KeyPressMsg:
		// Global keys
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Browser):
			m.openInBrowser()
			return m, nil
		}

		if m.loading || m.submitting || m.success {
			return m, nil
		}

		// Form navigation
		switch {
		case key.Matches(msg, m.keys.Enter):
			if !m.submitting {
				m.submitting = true
				m.setLogMessage("Creating install...", "info")
				return m, m.submitForm()
			}

		case key.Matches(msg, m.keys.Tab):
			m.nextInput()
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			m.prevInput()
			return m, nil

		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		default:
			if m.focusIndex == 1 {
				// Region field
				if msg.String() == "left" || msg.String() == "h" {
					m.regionIndex--
					if m.regionIndex < 0 {
						m.regionIndex = len(awsRegions) - 1
					}
					m.updateViewportContent()
				} else if msg.String() == "right" || msg.String() == "l" {
					m.regionIndex++
					if m.regionIndex >= len(awsRegions) {
						m.regionIndex = 0
					}
					m.updateViewportContent()
				}
			} else {
				// Text input field
				if inputIdx := m.focusIndexToInputIndex(m.focusIndex); inputIdx >= 0 {
					m.inputs[inputIdx], cmd = m.inputs[inputIdx].Update(msg)
					cmds = append(cmds, cmd)
					m.updateViewportContent()
				}
			}
		}

	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		if !m.loading && !m.submitting && !m.success {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func InstallCreatorApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	appID string,
) (string, error) {
	if !cfg.Interactive {
		return "", errors.New("interactive terminal required for install creation; use nuon installs create --name <name> --region <region> flags")
	}

	m := initialModel(ctx, cfg, api, appID)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running install creator: %v\n", err)
		os.Exit(1)
	}
	if fm, ok := finalModel.(model); ok {
		if fm.installID != "" {
			return fm.installID, nil
		}
	}
	return "", errors.New("unable to get install id")
}
