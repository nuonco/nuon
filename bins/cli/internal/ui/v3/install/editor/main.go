/*

An alt-screen TUI for editing an install's inputs, with a dynamically generated
form based on the app's input config and the install's current values.

*/

package editor

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
	installID string

	width  int
	height int

	// data
	inputConfig   *models.AppAppInputConfig
	install       *models.AppInstall
	currentInputs *models.AppInstallInputs

	// form state
	inputs           []textinput.Model
	inputMappings    []inputMapping
	focusIndex       int
	deployDependents bool

	// ui components
	viewport viewport.Model
	spinner  spinner.Model
	help     help.Model
	status   common.StatusBarRequest

	// field position tracking for scroll-into-view
	fieldEndLines map[int]int

	// state
	loading    bool
	submitting bool
	error      error
	success    bool
	workflowID string
	quitting   bool
	keys       keyMap
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	deployDependents bool,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	vp := viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(minRequiredHeight))
	vp.YPosition = 0

	m := model{
		ctx:              ctx,
		cfg:              cfg,
		api:              api,
		installID:        installID,
		deployDependents: deployDependents,
		viewport:         vp,
		spinner:          s,
		help:             help.New(),
		status:           common.StatusBarRequest{Message: "Loading install inputs..."},
		loading:          true,
		keys:             keys,
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
			m.setLogMessage(fmt.Sprintf("Error loading inputs: %s", msg.err), "error")
		} else {
			m.inputConfig = msg.inputConfig
			m.install = msg.install
			m.currentInputs = msg.currentInputs
			m.createFormInputs()
			m.setLogMessage("Edit the inputs and press Enter to save", "info")
		}
		return m, nil

	case inputsUpdatedMsg:
		m.submitting = false
		if msg.err != nil {
			m.error = msg.err
			m.setLogMessage(fmt.Sprintf("Error updating inputs: %s", msg.err), "error")
		} else {
			m.success = true
			if msg.inputs != nil {
				m.workflowID = msg.inputs.WorkflowID
			}
			m.setLogMessage("Inputs updated successfully", "success")
			return m, autoExitAfterDelay()
		}
		return m, nil

	case autoExitMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.PasteMsg:
		// Bracketed paste (e.g. cmd+v) arrives as a single PasteMsg, which the
		// textinput component does not handle itself. Insert it at the cursor
		// of the focused text field.
		if !m.loading && !m.submitting && !m.success && m.focusIndex < len(m.inputs) {
			m.insertAtCursor(m.focusIndex, msg.Content)
			m.updateViewportContent()
		}
		return m, nil

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
				m.setLogMessage("Updating inputs...", "info")
				return m, m.submitForm()
			}

		case key.Matches(msg, m.keys.Tab):
			m.nextInput()
			m.ensureFocusVisible()
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			m.prevInput()
			m.ensureFocusVisible()
			return m, nil

		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		default:
			if m.focusIndex == m.toggleIndex() {
				// Deploy dependents toggle
				switch msg.String() {
				case " ", "left", "right", "h", "l":
					m.deployDependents = !m.deployDependents
					m.updateViewportContent()
				}
			} else if m.focusIndex < len(m.inputs) {
				// Text input field
				m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
				m.updateViewportContent()
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

// EditInputsApp launches the install inputs editor TUI.
func EditInputsApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	installID string,
	deployDependents bool,
) error {
	if !cfg.Interactive {
		return errors.New("interactive terminal required for editing inputs; use `nuon installs inputs set key=value` instead")
	}

	m := initialModel(ctx, cfg, api, installID, deployDependents)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running inputs editor: %v\n", err)
		os.Exit(1)
	}
	return nil
}
