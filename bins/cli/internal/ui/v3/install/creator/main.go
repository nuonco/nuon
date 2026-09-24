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

	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"

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

// createStep is which screen the creator is showing. Group selection follows the
// form so the group's labels can be merged into the create request.
type createStep int

const (
	stepForm createStep = iota
	stepGroup
)

type model struct {
	// common/base
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// top level information
	appID        string
	name         string
	presetRegion string
	presetLabels map[string]string
	appBranchID  string

	width  int
	height int

	// data
	inputConfig   *models.AppAppInputConfig
	app           *models.AppApp
	cloudPlatform models.AppCloudPlatform

	// form state
	inputs              []textinput.Model
	focusIndex          int
	regionIndex         int
	inputMappings       []inputMapping
	nameCheckGeneration int
	nameChecked         string
	nameChecking        bool
	nameValidationErr   error

	// group selection, shown after the form when the branch has selectable groups
	step       createStep
	groups     []*models.AppAppBranchInstallGroup
	groupIndex int

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
	installID  string
	quitting   bool
	keys       keyMap
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	appID string,
	name string,
	region string,
	labels map[string]string,
	appBranchID string,
	inputConfig *models.AppAppInputConfig,
	groups []*models.AppAppBranchInstallGroup,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	vp := viewport.New(viewport.WithWidth(minRequiredWidth), viewport.WithHeight(minRequiredHeight))
	vp.YPosition = 0

	m := model{
		ctx:          ctx,
		cfg:          cfg,
		api:          api,
		appID:        appID,
		name:         name,
		presetRegion: region,
		presetLabels: labels,
		appBranchID:  appBranchID,
		inputConfig:  inputConfig,
		groups:       groups,
		viewport:     vp,
		spinner:      s,
		help:         help.New(),
		status:       common.StatusBarRequest{Message: "Loading app configuration..."},
		loading:      true,
		keys:         keys,
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
			m.cloudPlatform = msg.cloudPlatform
			m.createFormInputs()
			m.setLogMessage("Fill in the form and press Enter to create install", "info")
			return m, m.scheduleNameCheck()
		}
		return m, nil

	case nameCheckDebounceMsg:
		if msg.generation != m.nameCheckGeneration {
			return m, nil
		}
		return m, checkInstallNameCmd(m, msg.name, msg.generation)

	case nameCheckedMsg:
		if msg.generation != m.nameCheckGeneration {
			return m, nil
		}
		m.nameChecking = false
		m.nameChecked = msg.name
		switch {
		case msg.err != nil:
			m.nameValidationErr = fmt.Errorf("unable to check install name: %w", msg.err)
		case msg.exists:
			m.nameValidationErr = errDuplicateInstallName(msg.name)
		default:
			m.nameValidationErr = nil
			m.setLogMessage("Fill in the form and press Enter to create install", "info")
		}
		m.updateViewportContent()
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

		if m.step == stepGroup {
			return m.updateGroupStep(msg)
		}

		// Form navigation
		switch {
		case key.Matches(msg, m.keys.Enter):
			if !m.submitting {
				if err := m.validateForm(); err != nil {
					m.setLogMessage(err.Error(), "error")
					return m, nil
				}
				if len(m.groups) > 0 {
					m.step = stepGroup
					m.viewport.SetYOffset(0)
					m.setLogMessage("Select an install group, or skip to leave the install orphaned", "info")
					m.updateViewportContent()
					return m, nil
				}
				m.submitting = true
				m.setLogMessage("Creating install...", "info")
				return m, m.submitForm()
			}

		case key.Matches(msg, m.keys.Tab):
			if m.focusIndex == 0 {
				if err := m.validateName(); err != nil {
					m.setLogMessage(err.Error(), "error")
					return m, nil
				}
			}
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
			if m.needsRegion() && m.focusIndex == 1 {
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
					previousValue := m.inputs[inputIdx].Value()
					m.inputs[inputIdx], cmd = m.inputs[inputIdx].Update(msg)
					cmds = append(cmds, cmd)
					if inputIdx == 0 && m.inputs[inputIdx].Value() != previousValue {
						cmds = append(cmds, m.scheduleNameCheck())
					}
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

// updateGroupStep handles keys on the group selection screen. The last row is the
// skip option, which leaves the install without group labels.
func (m model) updateGroupStep(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.groupIndex > 0 {
			m.groupIndex--
			m.updateViewportContent()
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.groupIndex < len(m.groups) {
			m.groupIndex++
			m.updateViewportContent()
		}
		return m, nil

	case key.Matches(msg, m.keys.ShiftTab):
		m.step = stepForm
		m.viewport.SetYOffset(0)
		m.setLogMessage("Fill in the form and press Enter to continue", "info")
		m.updateViewportContent()
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if err := m.applyGroupSelection(); err != nil {
			m.setLogMessage(err.Error(), "error")
			return m, nil
		}
		m.submitting = true
		m.setLogMessage("Creating install...", "info")
		return m, m.submitForm()
	}

	return m, nil
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
	name string,
	region string,
	labels map[string]string,
	appBranchID string,
	inputConfig *models.AppAppInputConfig,
	groups []*models.AppAppBranchInstallGroup,
) (string, error) {
	if !cfg.Interactive {
		return "", errors.New("interactive terminal required for install creation; use nuon installs create --name <name> --region <region> flags")
	}

	m := initialModel(ctx, cfg, api, appID, name, region, labels, appBranchID, inputConfig, groups)
	p := teaprogram.NewProgram(m)

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
