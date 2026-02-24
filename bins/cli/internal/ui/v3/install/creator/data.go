package creator

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var awsRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}

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

		return configFetchedMsg{
			inputConfig: inputConfig,
			app:         app,
		}
	}
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

	// Focus the first input
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
		m.focusIndex = 0
	}

	m.updateViewportContent()
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
	return focusIdx - 1
}

func (m *model) nextInput() {
	if currentInputIdx := m.focusIndexToInputIndex(m.focusIndex); currentInputIdx >= 0 {
		m.inputs[currentInputIdx].Blur()
	}

	m.focusIndex++
	totalFields := len(m.inputs) + 1 // +1 for region field
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
	totalFields := len(m.inputs) + 1
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
