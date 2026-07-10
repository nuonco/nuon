package creator

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var awsRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"af-south-1",
	"ap-east-1", "ap-south-1", "ap-south-2",
	"ap-southeast-1", "ap-southeast-2", "ap-southeast-3", "ap-southeast-4",
	"ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
	"ca-central-1", "ca-west-1",
	"eu-central-1", "eu-central-2",
	"eu-west-1", "eu-west-2", "eu-west-3",
	"eu-south-1", "eu-south-2",
	"eu-north-1",
	"il-central-1",
	"me-south-1", "me-central-1",
	"sa-east-1",
	"us-gov-east-1", "us-gov-west-1",
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

		cloudPlatform := models.AppCloudPlatformAws
		if runnerCfg, err := m.api.GetAppRunnerLatestConfig(m.ctx, m.appID); err == nil && runnerCfg != nil && runnerCfg.CloudPlatform != "" {
			cloudPlatform = runnerCfg.CloudPlatform
		}

		return configFetchedMsg{
			inputConfig:   inputConfig,
			app:           app,
			cloudPlatform: cloudPlatform,
		}
	}
}

// needsRegion reports whether the install creation form should collect an AWS
// region. GCP and Azure installs have their region determined automatically
// from the stack output after provisioning.
func (m *model) needsRegion() bool {
	return m.cloudPlatform != models.AppCloudPlatformGcp && m.cloudPlatform != models.AppCloudPlatformAzure
}

// regionOffset is 1 when the region field is present (shifting focus indexes
// for the dynamic input fields that follow it), otherwise 0.
func (m *model) regionOffset() int {
	if m.needsRegion() {
		return 1
	}
	return 0
}

// fieldPrefilled reports whether the field at focusIdx was already supplied
// via a CLI flag (--name / --region), and so doesn't need user input.
func (m *model) fieldPrefilled(focusIdx int) bool {
	if focusIdx == 0 {
		return m.name != ""
	}
	if m.needsRegion() && focusIdx == 1 {
		return m.presetRegion != ""
	}
	return false
}

func (m *model) createFormInputs() {
	m.inputs = make([]textinput.Model, 0)
	m.inputMappings = make([]inputMapping, 0)

	// 1. Name field
	nameInput := textinput.New()
	nameInput.Placeholder = "my-install"
	nameInput.CharLimit = 100
	nameInput.SetWidth(50)
	nameInput.Prompt = ""
	if m.name != "" {
		nameInput.SetValue(m.name)
	}
	m.inputs = append(m.inputs, nameInput)
	m.inputMappings = append(m.inputMappings, inputMapping{
		name:        "name",
		displayName: "Install Name",
		description: "Name for this installation",
		required:    true,
	})

	// 2. Region is handled separately with regionIndex
	if m.presetRegion != "" {
		for i, r := range awsRegions {
			if r == m.presetRegion {
				m.regionIndex = i
				break
			}
		}
	}

	// 3. Dynamic inputs from app config, organized by groups
	if m.inputConfig != nil && m.inputConfig.InputGroups != nil {
		for _, group := range m.inputConfig.InputGroups {
			if group.AppInputs == nil {
				continue
			}

			for _, input := range group.AppInputs {
				if input.Internal {
					continue
				}

				ti := textinput.New()
				ti.Placeholder = fmt.Sprintf("Enter %s", input.DisplayName)
				ti.CharLimit = 500
				ti.SetWidth(50)
				ti.Prompt = ""

				if input.Default != "" {
					ti.SetValue(input.Default)
				}

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

	// Focus the first field that wasn't already pre-filled via --name/--region
	// flags, stopping at the last field so there's always something focused.
	if len(m.inputs) > 0 {
		m.focusIndex = 0
		totalFields := len(m.inputs) + m.regionOffset()
		for m.focusIndex < totalFields-1 && m.fieldPrefilled(m.focusIndex) {
			m.nextInput()
		}
		if newInputIdx := m.focusIndexToInputIndex(m.focusIndex); newInputIdx >= 0 {
			m.inputs[newInputIdx].Focus()
		}
	}

	m.updateViewportContent()
}

// focusIndexToInputIndex converts a focusIndex (which includes the region
// field at index 1, when present) to an index in the m.inputs array. Returns
// -1 if focusIndex points to the region field.
func (m *model) focusIndexToInputIndex(focusIdx int) int {
	if focusIdx == 0 {
		return 0 // name field
	}
	if m.needsRegion() && focusIdx == 1 {
		return -1 // region field, not in inputs array
	}
	return focusIdx - m.regionOffset()
}

func (m *model) nextInput() {
	if currentInputIdx := m.focusIndexToInputIndex(m.focusIndex); currentInputIdx >= 0 {
		m.inputs[currentInputIdx].Blur()
	}

	m.focusIndex++
	totalFields := len(m.inputs) + m.regionOffset()
	if m.focusIndex >= totalFields {
		m.focusIndex = 0
	}

	if newInputIdx := m.focusIndexToInputIndex(m.focusIndex); newInputIdx >= 0 {
		m.inputs[newInputIdx].Focus()
	}

	m.updateViewportContent()
}

func (m *model) prevInput() {
	if currentInputIdx := m.focusIndexToInputIndex(m.focusIndex); currentInputIdx >= 0 {
		m.inputs[currentInputIdx].Blur()
	}

	m.focusIndex--
	totalFields := len(m.inputs) + m.regionOffset()
	if m.focusIndex < 0 {
		m.focusIndex = totalFields - 1
	}

	if newInputIdx := m.focusIndexToInputIndex(m.focusIndex); newInputIdx >= 0 {
		m.inputs[newInputIdx].Focus()
	}

	m.updateViewportContent()
}

func (m *model) validateForm() error {
	if strings.TrimSpace(m.inputs[0].Value()) == "" {
		return fmt.Errorf("install name is required")
	}

	for i, mapping := range m.inputMappings {
		if i == 0 {
			continue
		}
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

		inputsMap := make(map[string]string)
		for i, mapping := range m.inputMappings {
			if i == 0 {
				continue
			}
			value := strings.TrimSpace(m.inputs[i].Value())
			if value != "" {
				inputsMap[mapping.name] = value
			}
		}

		name := strings.TrimSpace(m.inputs[0].Value())

		req := &models.ServiceCreateInstallRequest{
			Name:   &name,
			Inputs: inputsMap,
			Labels: m.presetLabels,
		}
		switch m.cloudPlatform {
		case models.AppCloudPlatformGcp:
			req.GcpAccount = &models.ServiceCreateInstallRequestGcpAccount{}
		case models.AppCloudPlatformAzure:
			req.AzureAccount = &models.ServiceCreateInstallRequestAzureAccount{}
		default:
			req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{
				Region: awsRegions[m.regionIndex],
			}
		}

		install, err := m.api.CreateInstall(m.ctx, m.appID, req)

		if err != nil {
			return installCreatedMsg{err: err}
		}

		return installCreatedMsg{install: install}
	}
}
