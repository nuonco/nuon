package run

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// getStepIDToNameMap creates a map from step ID to step name
// by matching config steps with run steps
func (m *Model) getStepIDToNameMap() map[string]string {
	stepIDToName := make(map[string]string)

	if m.run == nil || m.run.Config == nil || m.run.Config.Steps == nil {
		return stepIDToName
	}

	// Build map from config steps (which have the names)
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
	// header
	sections = append(sections, fmt.Sprintf("%d Steps", len(m.run.Steps)))

	// Get map of step IDs to names
	stepIDToName := m.getStepIDToNameMap()

	// steps
	stepStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(m.sidebar.Width - 2)
	for _, step := range m.run.Steps {
		status := styles.GetStatusStyle(models.AppStatus(step.Status)).Render(fmt.Sprintf("[%s]", step.Status))

		// Get step name from map, fallback to ID if not found
		stepName := stepIDToName[step.StepID]
		if stepName == "" {
			stepName = step.StepID
		}

		stepSection := stepStyle.Render(fmt.Sprintf("%s %s", status, stepName))
		sections = append(sections, stepSection)
	}

	sections = append(sections, "")
	if m.run.Outputs != nil {
		sections = append(sections, "outputs")
		// Convert the generic interface to a map
		outputsMap, ok := m.run.Outputs.(map[string]any)
		if ok {
			// Sort keys alphabetically to prevent jitter
			keys := make([]string, 0, len(outputsMap))
			for k := range outputsMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// Render each key-value pair with special handling for "steps"
			for _, key := range keys {
				value := outputsMap[key]

				// Special handling for "steps" array - expand it
				if key == "steps" {
					sections = append(sections, fmt.Sprintf("  %s: [", key))
					if stepsArray, ok := value.([]any); ok {
						for i, step := range stepsArray {
							// Keep step content unprettified
							stepJSON, err := json.Marshal(step)
							if err != nil {
								sections = append(sections, "    <error>")
								continue
							}
							// Add comma for all but last element
							if i < len(stepsArray)-1 {
								sections = append(sections, fmt.Sprintf("    %s,", string(stepJSON)))
							} else {
								sections = append(sections, fmt.Sprintf("    %s", string(stepJSON)))
							}
						}
					}
					sections = append(sections, "  ]")
				} else {
					// For other keys, use pretty-printed JSON
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
