/*

An alt-screen TUI for creating installs with dynamic form generation based on app inputs.

*/

package creator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	maxWidth              = 80
	minRequiredWidth  int = 60
	minRequiredHeight int = 16
)

var awsRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}

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
	inputs        []textinput.Model // dynamic inputs based on app config
	focusIndex    int
	regionIndex   int // index in awsRegions
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

type inputMapping struct {
	name             string
	displayName      string
	description      string
	inputType        string
	required         bool
	sensitive        bool
	groupName        string
	groupDescription string
	groupID          string
}

type configFetchedMsg struct {
	inputConfig *models.AppAppInputConfig
	app         *models.AppApp
	err         error
}

type installCreatedMsg struct {
	install *models.AppInstall
	err     error
}

type autoExitMsg struct{}

func autoExitAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return autoExitMsg{}
	})
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

	vp := viewport.New(minRequiredWidth, minRequiredHeight)
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

func fetchConfigCmd(m model) tea.Cmd {
	return func() tea.Msg {
		inputConfig, err := m.api.GetAppInputLatestConfig(m.ctx, m.appID)
		if err != nil {
			return configFetchedMsg{err: err}
		}

		app, err := m.api.GetApp(m.ctx, m.appID)
		if err != nil {
			return configFetchedMsg{err: err}
		}

		return configFetchedMsg{
			inputConfig: inputConfig,
			app:         app,
		}
	}
}

func (m *model) createFormInputs() {
	// Create inputs list: name, region, then dynamic inputs
	m.inputs = make([]textinput.Model, 0)
	m.inputMappings = make([]inputMapping, 0)

	// 1. Name field
	nameInput := textinput.New()
	nameInput.Placeholder = "my-install"
	nameInput.CharLimit = 100
	nameInput.Width = 50
	nameInput.Prompt = ""
	m.inputs = append(m.inputs, nameInput)
	m.inputMappings = append(m.inputMappings, inputMapping{
		name:        "name",
		displayName: "Install Name",
		description: "Name for this installation",
		required:    true,
	})

	// 2. Region is handled separately with regionIndex

	// 3. Dynamic inputs from app config, organized by groups
	if m.inputConfig != nil && m.inputConfig.InputGroups != nil {
		for _, group := range m.inputConfig.InputGroups {
			if group.AppInputs == nil {
				continue
			}

			for _, input := range group.AppInputs {
				// Skip internal inputs
				if input.Internal {
					continue
				}

				ti := textinput.New()
				ti.Placeholder = fmt.Sprintf("Enter %s", input.DisplayName)
				ti.CharLimit = 500
				ti.Width = 50
				ti.Prompt = ""

				// Set default value if provided
				if input.Default != "" {
					ti.SetValue(input.Default)
				}

				// Handle sensitive inputs
				if input.Sensitive {
					ti.EchoMode = textinput.EchoPassword
					ti.EchoCharacter = '•'
				}

				m.inputs = append(m.inputs, ti)
				m.inputMappings = append(m.inputMappings, inputMapping{
					name:             input.Name,
					displayName:      input.DisplayName,
					description:      input.Description,
					inputType:        input.Type,
					required:         input.Required,
					sensitive:        input.Sensitive,
					groupName:        group.DisplayName,
					groupDescription: group.Description,
					groupID:          group.ID,
				})
			}
		}
	}

	// Focus the first input
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
		m.focusIndex = 0
	}

	// Update viewport with the form content
	m.updateViewportContent()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchConfigCmd(m),
		m.spinner.Tick,
	)
}

// focusIndexToInputIndex converts a focusIndex (which includes region at index 1)
// to an index in the m.inputs array. Returns -1 if focusIndex points to region field.
func (m *model) focusIndexToInputIndex(focusIdx int) int {
	if focusIdx == 1 {
		return -1 // region field, not in inputs array
	}
	if focusIdx == 0 {
		return 0 // name field
	}
	// Dynamic inputs: focusIdx 2+ maps to inputs[1+]
	return focusIdx - 1
}

func (m *model) nextInput() {
	// Blur current input if it's not the region field
	if currentInputIdx := m.focusIndexToInputIndex(m.focusIndex); currentInputIdx >= 0 {
		m.inputs[currentInputIdx].Blur()
	}

	// Move to next field
	m.focusIndex++
	// +1 for region field
	totalFields := len(m.inputs) + 1
	if m.focusIndex >= totalFields {
		m.focusIndex = 0
	}

	// Focus new input if it's not the region field
	if newInputIdx := m.focusIndexToInputIndex(m.focusIndex); newInputIdx >= 0 {
		m.inputs[newInputIdx].Focus()
	}

	// Update viewport to reflect focus change
	m.updateViewportContent()
}

func (m *model) prevInput() {
	// Blur current input if it's not the region field
	if currentInputIdx := m.focusIndexToInputIndex(m.focusIndex); currentInputIdx >= 0 {
		m.inputs[currentInputIdx].Blur()
	}

	// Move to previous field
	m.focusIndex--
	totalFields := len(m.inputs) + 1
	if m.focusIndex < 0 {
		m.focusIndex = totalFields - 1
	}

	// Focus new input if it's not the region field
	if newInputIdx := m.focusIndexToInputIndex(m.focusIndex); newInputIdx >= 0 {
		m.inputs[newInputIdx].Focus()
	}

	// Update viewport to reflect focus change
	m.updateViewportContent()
}

// updateViewportContent builds the form content and sets it in the viewport
// This should be called whenever the form content changes (not in View())
func (m *model) updateViewportContent() {
	width := common.Min(m.width, maxWidth) - 4
	sections := []string{}

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		Padding(1, 0)

	labelStyle := lipgloss.NewStyle().
		Foreground(styles.TextColor).
		Bold(true)

	descStyle := styles.TextDim.Italic(true)

	focusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderActiveColor).
		Padding(0, 1)

	blurredStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderInactiveColor).
		Padding(0, 1)

	// Title
	title := titleStyle.Render("Create Install")
	if m.app != nil {
		title = titleStyle.Render(fmt.Sprintf("Create Install for %s", m.app.Name))
	}
	sections = append(sections, title)

	// Render name field first
	if len(m.inputMappings) > 0 {
		mapping := m.inputMappings[0]
		label := labelStyle.Render(mapping.displayName)
		if mapping.required {
			label += styles.TextError.Render(" *")
		}
		sections = append(sections, label)

		if mapping.description != "" {
			sections = append(sections, descStyle.Render(mapping.description))
		}

		fieldContent := m.inputs[0].View()
		if m.focusIndex == 0 {
			sections = append(sections, focusedStyle.Render(fieldContent))
		} else {
			sections = append(sections, blurredStyle.Render(fieldContent))
		}
	}

	// Render region field (focusIndex 1)
	sections = append(sections, labelStyle.Render("AWS Region"))
	sections = append(sections, styles.TextError.Render(" *"))
	sections = append(sections, descStyle.Render("AWS region for the installation (use left/right arrows to change)"))

	regionDisplay := fmt.Sprintf("  %s  ", awsRegions[m.regionIndex])
	if m.focusIndex == 1 {
		regionDisplay = focusedStyle.Render(regionDisplay)
	} else {
		regionDisplay = blurredStyle.Render(regionDisplay)
	}
	sections = append(sections, regionDisplay)
	sections = append(sections, "\n")

	// Group header style
	groupHeaderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderForeground(styles.SubtleColor).
		Width(width).
		Padding(0, 1)
	groupTitleStyle := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Bold(true)

	groupInputsStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1)

	// Render dynamic input fields (focusIndex 2+), grouped by their groups
	lastGroupID := ""
	for i := 1; i < len(m.inputMappings); i++ {
		mapping := m.inputMappings[i]

		// Render group header when group changes
		if mapping.groupID != "" && mapping.groupID != lastGroupID {
			if lastGroupID != "" {
				// Add spacing between groups (not for first group)
				sections = append(sections, "\n")
			}
			groupTitle := lipgloss.JoinVertical(
				lipgloss.Top,
				groupTitleStyle.Render(mapping.groupName),
				styles.TextDim.Render(mapping.groupDescription),
			)

			sections = append(sections, groupHeaderStyle.Render(groupTitle))
			lastGroupID = mapping.groupID
		}

		// Render input field
		inputSections := []string{}

		label := labelStyle.Render(mapping.displayName)
		if mapping.required {
			label += styles.TextError.Render(" *")
		}
		inputSections = append(inputSections, label)

		if mapping.description != "" {
			inputSections = append(inputSections, styles.TextAccent.Render(mapping.description))
		}

		fieldContent := m.inputs[i].View()
		// Dynamic inputs have focusIndex = i + 1 (because region is at 1)
		if m.focusIndex == i+1 {
			inputSections = append(inputSections, focusedStyle.Render(fieldContent))
		} else {
			inputSections = append(inputSections, blurredStyle.Render(fieldContent))
		}
		sections = append(sections,
			groupInputsStyle.Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					inputSections...,
				),
			),
		)
	}

	// Set the viewport content
	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, sections...))
}

func (m *model) validateForm() error {
	// Check name
	if strings.TrimSpace(m.inputs[0].Value()) == "" {
		return fmt.Errorf("install name is required")
	}

	// Check required fields
	for i, mapping := range m.inputMappings {
		if i == 0 {
			continue // skip name, already checked
		}
		// inputMappings and inputs arrays are parallel, so same index
		if mapping.required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			return fmt.Errorf("%s is required", mapping.displayName)
		}
	}

	return nil
}

func (m *model) submitForm() tea.Cmd {
	return func() tea.Msg {
		if err := m.validateForm(); err != nil {
			return installCreatedMsg{err: err}
		}

		// Build inputs map
		inputsMap := make(map[string]string)
		for i, mapping := range m.inputMappings {
			if i == 0 {
				continue // skip name field
			}
			// inputMappings and inputs arrays are parallel, so same index
			value := strings.TrimSpace(m.inputs[i].Value())
			if value != "" {
				inputsMap[mapping.name] = value
			}
		}

		name := strings.TrimSpace(m.inputs[0].Value())
		region := awsRegions[m.regionIndex]

		install, _, err := m.api.CreateInstall(m.ctx, m.appID, &models.ServiceCreateInstallRequest{
			Name: &name,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region: region,
			},
			Inputs: inputsMap,
		})

		if err != nil {
			return installCreatedMsg{err: err}
		}

		return installCreatedMsg{install: install}
	}
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

		// Calculate viewport height (subtract help and status space)
		helpHeight := lipgloss.Height(m.help.View(m.keys))
		statusHeight := 3 // space for status message
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - helpHeight - statusHeight - 2

		return m, nil

	case tea.KeyMsg:
		// Handle global keys first
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

		// Don't process input if loading or submitting or success
		if m.loading || m.submitting || m.success {
			return m, nil
		}

		// Handle form navigation
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
			// Viewport scrolling
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		default:
			// Handle input for current field
			if m.focusIndex == 1 {
				// Region field - use left/right arrow keys to cycle
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
				// Text input field - pass through keystrokes
				if inputIdx := m.focusIndexToInputIndex(m.focusIndex); inputIdx >= 0 {
					m.inputs[inputIdx], cmd = m.inputs[inputIdx].Update(msg)
					cmds = append(cmds, cmd)
					m.updateViewportContent()
				}
			}
		}

	default:
		// Update spinner
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Also update viewport for mouse events
		if !m.loading && !m.submitting && !m.success {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 {
		return ""
	}

	if m.width < minRequiredWidth || m.height < minRequiredHeight {
		return common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Level:   "warning",
			Content: lipgloss.JoinVertical(
				lipgloss.Center,
				"  This screen is too small  ",
				fmt.Sprintf("Minimum dimensions %d x %d", minRequiredWidth, minRequiredHeight),
			),
		})
	}

	if m.loading {
		return common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Content: fmt.Sprintf("Loading %s", m.spinner.View()),
			Level:   "info",
		})
	}

	if m.error != nil && m.inputConfig == nil {
		return common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Content: fmt.Sprintf("Error: %s", m.error.Error()),
			Level:   "error",
		})
	}

	if m.success {
		cfg, _ := m.api.GetCLIConfig(m.ctx)
		url := ""
		if cfg != nil {
			url = fmt.Sprintf("\n\n%s/%s/installs/%s", cfg.DashboardURL, m.cfg.OrgID, m.installID)
		}

		return common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Content: fmt.Sprintf("Install created successfully!\n\nInstall ID: %s%s\n\nExiting in 3 seconds...", m.installID, url),
			Level:   "success",
		})
	}

	// Build the final view with viewport, status, and help (content already set elsewhere)
	var finalView strings.Builder

	// Add viewport
	finalView.WriteString(m.viewport.View())
	finalView.WriteString("\n")

	// Status message
	if m.status.Message != "" {
		statusStyle := lipgloss.NewStyle()
		switch m.status.Level {
		case "error":
			statusStyle = statusStyle.Foreground(styles.ErrorColor)
		case "success":
			statusStyle = statusStyle.Foreground(styles.SuccessColor)
		case "warning":
			statusStyle = statusStyle.Foreground(styles.WarningColor)
		default:
			statusStyle = statusStyle.Foreground(styles.InfoColor)
		}

		finalView.WriteString("\n")
		finalView.WriteString(statusStyle.Render(m.status.Message))
	}

	// Help
	helpView := m.help.View(m.keys)
	finalView.WriteString("\n")
	finalView.WriteString(lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(helpView))

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(finalView.String())
}

func InstallCreatorApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	appID string,
) (string, error) {
	m := initialModel(ctx, cfg, api, appID)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

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
