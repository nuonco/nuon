package run

import (
	"encoding/json"
	"fmt"
	"sort"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (m *Model) getStepIDToNameMap() map[string]string {
	stepIDToName := make(map[string]string)

	if m.run == nil || m.run.Config == nil || m.run.Config.Steps == nil {
		return stepIDToName
	}

	for _, configStep := range m.run.Config.Steps {
		if configStep != nil && configStep.ID != "" {
			name := configStep.Name
			if name == "" {
				name = "Unnamed Step"
			}
			stepIDToName[configStep.ID] = name
		}
	}

	return stepIDToName
}

func (m *Model) setSidebarContent() string {
	if m.run == nil {
		m.sidebar.SetContent(" ... ")
		return " ... "
	}
	sections := []string{}
	sections = append(sections, fmt.Sprintf("%d Steps", len(m.run.Steps)))

	stepIDToName := m.getStepIDToNameMap()

	stepStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(m.sidebar.Width() - 2)
	for _, step := range m.run.Steps {
		status := styles.GetStatusStyle(models.AppStatus(step.Status)).Render(fmt.Sprintf("[%s]", step.Status))

		stepID := step.StepID
		stepName := stepIDToName[stepID]
		if stepName == "" {
			stepName = stepID
		}

		stepSection := stepStyle.Render(fmt.Sprintf("%s %s", status, stepName))
		sections = append(sections, stepSection)
	}

	sections = append(sections, "")
	if m.run.Outputs != nil {
		sections = append(sections, "outputs")
		outputsMap, ok := m.run.Outputs.(map[string]any)
		if ok {
			keys := make([]string, 0, len(outputsMap))
			for k := range outputsMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				value := outputsMap[key]

				if key == "steps" {
					sections = append(sections, fmt.Sprintf("  %s: [", key))
					if stepsArray, ok := value.([]any); ok {
						for i, step := range stepsArray {
							stepJSON, err := json.Marshal(step)
							if err != nil {
								sections = append(sections, "    <error>")
								continue
							}
							if i < len(stepsArray)-1 {
								sections = append(sections, fmt.Sprintf("    %s,", string(stepJSON)))
							} else {
								sections = append(sections, fmt.Sprintf("    %s", string(stepJSON)))
							}
						}
					}
					sections = append(sections, "  ]")
				} else {
					valueJSON, err := json.MarshalIndent(value, "  ", "  ")
					if err != nil {
						sections = append(sections, fmt.Sprintf("  %s: <error>", key))
						continue
					}
					sections = append(sections, fmt.Sprintf("  %s: %s", key, string(valueJSON)))
				}
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Top, sections...)
	m.sidebar.SetContent(content)
	return content
}
