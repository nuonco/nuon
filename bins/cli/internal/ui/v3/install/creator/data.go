package creator

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/nuonco/nuon/bins/cli/internal/installcreate"
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

const installNameDebounce = 300 * time.Millisecond

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
		inputConfig := m.inputConfig
		if inputConfig == nil {
			var err error
			inputConfig, err = installcreate.ResolveInputConfig(m.ctx, m.api, m.appID, m.appBranchID)
			if err != nil {
				return configFetchedMsg{err: err}
			}
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

func (m *model) needsRegion() bool {
	return m.cloudPlatform != models.AppCloudPlatformGcp && m.cloudPlatform != models.AppCloudPlatformAzure
}

func (m *model) regionOffset() int {
	if m.needsRegion() {
		return 1
	}
	return 0
}

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

	if m.presetRegion != "" {
		for i, r := range awsRegions {
			if r == m.presetRegion {
				m.regionIndex = i
				break
			}
		}
	}

	seenInputs := make(map[string]struct{})
	inputKey := func(input *models.AppAppInput) string {
		if input.ID != "" {
			return input.ID
		}
		return input.Name
	}
	if m.inputConfig != nil && m.inputConfig.InputGroups != nil {
		for _, group := range m.inputConfig.InputGroups {
			if group == nil || group.AppInputs == nil {
				continue
			}

			for _, input := range group.AppInputs {
				if input == nil || input.Internal {
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
				seenInputs[inputKey(input)] = struct{}{}
			}
		}
	}

	if m.inputConfig != nil {
		for _, input := range m.inputConfig.Inputs {
			if input == nil || input.Internal {
				continue
			}
			if _, ok := seenInputs[inputKey(input)]; ok {
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
				name:        input.Name,
				displayName: input.DisplayName,
				description: input.Description,
				inputType:   input.Type,
				required:    input.Required,
				sensitive:   input.Sensitive,
			})
		}
	}

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

func (m *model) focusIndexToInputIndex(focusIdx int) int {
	if focusIdx == 0 {
		return 0
	}
	if m.needsRegion() && focusIdx == 1 {
		return -1
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
	if err := m.validateName(); err != nil {
		return err
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

func (m *model) validateName() error {
	name := strings.TrimSpace(m.inputs[0].Value())
	if name == "" {
		return fmt.Errorf("install name is required")
	}
	if m.nameChecking || m.nameChecked != name {
		return fmt.Errorf("checking whether install name %q is available", name)
	}
	return m.nameValidationErr
}

func errDuplicateInstallName(name string) error {
	return fmt.Errorf("an install named %q already exists", name)
}

func (m *model) scheduleNameCheck() tea.Cmd {
	name := strings.TrimSpace(m.inputs[0].Value())
	m.nameCheckGeneration++
	generation := m.nameCheckGeneration
	m.nameChecked = ""
	m.nameValidationErr = nil
	m.nameChecking = name != ""

	if name == "" {
		return nil
	}
	m.setLogMessage("Checking install name availability...", "info")
	return tea.Tick(installNameDebounce, func(time.Time) tea.Msg {
		return nameCheckDebounceMsg{name: name, generation: generation}
	})
}

func checkInstallNameCmd(m model, name string, generation int) tea.Cmd {
	return func() tea.Msg {
		const pageSize = 100
		query := &models.GetPaginatedQuery{Limit: pageSize, Q: name}
		for {
			installs, hasNext, err := m.api.GetAppInstalls(m.ctx, m.appID, query)
			if err != nil {
				return nameCheckedMsg{name: name, generation: generation, err: err}
			}
			for _, install := range installs {
				if install != nil && install.Name == name {
					return nameCheckedMsg{name: name, generation: generation, exists: true}
				}
			}
			if !hasNext || len(installs) == 0 {
				return nameCheckedMsg{name: name, generation: generation}
			}
			query.Offset += len(installs)
		}
	}
}

func (m *model) selectedGroup() *models.AppAppBranchInstallGroup {
	if m.groupIndex < 0 || m.groupIndex >= len(m.groups) {
		return nil
	}
	return m.groups[m.groupIndex]
}

func (m *model) applyGroupSelection() error {
	group := m.selectedGroup()
	if group == nil {
		return nil
	}

	groupLabels, err := installcreate.GroupLabels(group)
	if err != nil {
		return err
	}

	merged := make(map[string]string, len(m.presetLabels)+len(groupLabels))
	for key, value := range m.presetLabels {
		merged[key] = value
	}
	if err := installcreate.MergeGroupLabels(merged, groupLabels); err != nil {
		return err
	}

	m.presetLabels = merged
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
			Name:        &name,
			Inputs:      inputsMap,
			Labels:      m.presetLabels,
			AppBranchID: m.appBranchID,
		}
		switch m.cloudPlatform {
		case models.AppCloudPlatformGcp:
			req.GcpAccount = &models.HelpersCreateInstallGCPAccountParams{}
		case models.AppCloudPlatformAzure:
			req.AzureAccount = &models.HelpersCreateInstallAzureAccountParams{}
		default:
			req.AwsAccount = &models.HelpersCreateInstallAWSAccountParams{
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
