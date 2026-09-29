package editor

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

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
	defaultValue     string
	currentValue     string
}

func fetchConfigCmd(m model) tea.Cmd {
	return func() tea.Msg {
		install, err := m.api.GetInstall(m.ctx, m.installID)
		if err != nil {
			return configFetchedMsg{err: err}
		}

		inputConfig, err := m.api.GetAppInputLatestConfig(m.ctx, install.AppID)
		if err != nil {
			return configFetchedMsg{err: err}
		}

		currentInputs, err := m.api.GetInstallCurrentInputs(m.ctx, m.installID)
		if err != nil {
			return configFetchedMsg{err: err}
		}

		return configFetchedMsg{
			inputConfig:   inputConfig,
			install:       install,
			currentInputs: currentInputs,
		}
	}
}

func (im inputMapping) label() string {
	if im.displayName != "" {
		return im.displayName
	}
	return im.name
}

func (m *model) createFormInputs() {
	m.inputs = make([]textinput.Model, 0)
	m.inputMappings = make([]inputMapping, 0)

	current := map[string]string{}
	if m.currentInputs != nil && m.currentInputs.Values != nil {
		current = m.currentInputs.Values
	}

	addInput := func(input *models.AppAppInput, group *models.AppAppInputGroup) {
		ti := textinput.New()
		ti.CharLimit = 500
		ti.SetWidth(50)
		ti.Prompt = ""

		curVal := current[input.Name]
		if input.Sensitive {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
			ti.Placeholder = "leave unchanged"
			if curVal != "" {
				ti.SetValue(curVal)
			}
		} else {
			ti.Placeholder = fmt.Sprintf("Enter %s", input.Name)
			switch {
			case curVal != "":
				ti.SetValue(curVal)
			case input.Default != "":
				ti.SetValue(input.Default)
			}
		}

		mapping := inputMapping{
			name:         input.Name,
			displayName:  input.DisplayName,
			description:  input.Description,
			inputType:    input.Type,
			required:     input.Required,
			sensitive:    input.Sensitive,
			defaultValue: input.Default,
			currentValue: curVal,
		}
		if group != nil {
			mapping.groupName = group.DisplayName
			mapping.groupDescription = group.Description
			mapping.groupID = group.ID
		}

		m.inputs = append(m.inputs, ti)
		m.inputMappings = append(m.inputMappings, mapping)
	}

	if m.inputConfig != nil && len(m.inputConfig.InputGroups) > 0 {
		for _, group := range m.inputConfig.InputGroups {
			for _, input := range group.AppInputs {
				addInput(input, group)
			}
		}
	} else if m.inputConfig != nil {
		for _, input := range m.inputConfig.Inputs {
			addInput(input, nil)
		}
	}

	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
		m.focusIndex = 0
	}

	m.updateViewportContent()
}

func (m *model) toggleIndex() int {
	return len(m.inputs)
}

func (m *model) totalFields() int {
	return len(m.inputs) + 1
}

func (m *model) nextInput() {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}

	m.focusIndex++
	if m.focusIndex >= m.totalFields() {
		m.focusIndex = 0
	}

	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}

	m.updateViewportContent()
}

func (m *model) prevInput() {
	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Blur()
	}

	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = m.totalFields() - 1
	}

	if m.focusIndex < len(m.inputs) {
		m.inputs[m.focusIndex].Focus()
	}

	m.updateViewportContent()
}

func (m *model) insertAtCursor(idx int, content string) {
	if idx < 0 || idx >= len(m.inputs) {
		return
	}
	ti := &m.inputs[idx]
	runes := []rune(ti.Value())
	pos := ti.Position()
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	inserted := []rune(content)
	newRunes := make([]rune, 0, len(runes)+len(inserted))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, inserted...)
	newRunes = append(newRunes, runes[pos:]...)
	ti.SetValue(string(newRunes))
	ti.SetCursor(pos + len(inserted))
}

func (m *model) validateForm() error {
	for i, mapping := range m.inputMappings {
		if mapping.required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			return fmt.Errorf("%s is required", mapping.label())
		}
	}
	return nil
}

func (m *model) submitForm() tea.Cmd {
	return func() tea.Msg {
		if err := m.validateForm(); err != nil {
			return inputsUpdatedMsg{err: err}
		}

		merged := make(map[string]string)
		if m.currentInputs != nil && m.currentInputs.Values != nil {
			for k, v := range m.currentInputs.Values {
				merged[k] = v
			}
		}
		for i, mapping := range m.inputMappings {
			merged[mapping.name] = m.inputs[i].Value()
		}

		deployDependents := m.deployDependents
		resp, err := m.api.UpdateInstallInputs(m.ctx, m.installID, &models.ServiceUpdateInstallInputsRequest{
			Inputs:           merged,
			DeployDependents: &deployDependents,
		})
		if err != nil {
			return inputsUpdatedMsg{err: err}
		}

		return inputsUpdatedMsg{inputs: resp}
	}
}
