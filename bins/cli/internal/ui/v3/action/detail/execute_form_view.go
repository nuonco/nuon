package detail

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m Model) renderExecuteForm() string {
	// Return the viewport view which handles scrolling and height constraints
	return m.formViewport.View()
}

func (m Model) renderExecuteFormContent() string {
	if m.formError != nil {
		req := common.FullPageDialogRequest{
			Width:   m.runsList.Width(),
			Height:  m.runsList.Height(),
			Content: styles.TextError.Padding(1).Width(m.runsList.Width() - 8).Render(fmt.Sprintf("%s", m.formError)),
			Level:   "error",
		}
		return common.FullPageDialog(req)
	}

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
	actionName := ""
	if m.installActionWorkflow != nil && m.installActionWorkflow.ActionWorkflow != nil {
		actionName = m.installActionWorkflow.ActionWorkflow.Name
	}
	title := titleStyle.Render(fmt.Sprintf("Execute: %s", actionName))
	sections = append(sections, title)

	if len(m.formMappings) == 0 {
		sections = append(sections, "")
		sections = append(sections, descStyle.Render("No environment variables configured."))
		sections = append(sections, descStyle.Render("Press Enter to execute without inputs."))
	} else {
		sections = append(sections, "")
		sections = append(sections, descStyle.Render("Environment variables:"))
		sections = append(sections, "")

		// Render each input field
		for i, mapping := range m.formMappings {
			label := labelStyle.Render(mapping.name)
			sections = append(sections, label)

			if mapping.value != "" {
				sections = append(sections, descStyle.Render(fmt.Sprintf("Default: %s", mapping.value)))
			}

			fieldContent := m.formInputs[i].View()
			if m.formFocusIndex == i {
				sections = append(sections, focusedStyle.Render(fieldContent))
			} else {
				sections = append(sections, blurredStyle.Render(fieldContent))
			}

			sections = append(sections, "")
		}
	}

	// Help text
	sections = append(sections, "")
	helpStyle := styles.TextDim.Italic(true)
	sections = append(sections, helpStyle.Render("Tab/Shift+Tab: navigate | Enter: execute | Esc: cancel"))

	content := lipgloss.JoinVertical(lipgloss.Top, sections...)

	// Apply minimal padding
	contentStyle := lipgloss.NewStyle().Padding(1, 2)

	return contentStyle.Render(content)
}
