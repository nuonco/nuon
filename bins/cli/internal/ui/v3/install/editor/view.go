package editor

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

func (m model) viewContent() string {
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
		content := "Inputs updated successfully!"
		if m.workflowID != "" {
			content += fmt.Sprintf("\n\nDeploy workflow: %s", m.workflowID)
		}
		content += "\n\nExiting in 3 seconds..."

		return common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Content: content,
			Level:   "success",
		})
	}

	var finalView strings.Builder

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

// updateViewportContent builds the form content and sets it in the viewport.
func (m *model) updateViewportContent() {
	width := min(m.width, maxWidth) - 4

	// fieldLines tracks the ending line (exclusive) for each focusIndex.
	fieldLines := map[int]int{}
	lineCount := 0

	appendSection := func(sections []string, s string) []string {
		lineCount += strings.Count(s, "\n") + 1
		return append(sections, s)
	}

	sections := []string{}

	// Title
	title := titleStyle.Render("Edit Inputs")
	if m.install != nil {
		title = titleStyle.Render(fmt.Sprintf("Edit Inputs for %s", m.install.Name))
	}
	sections = appendSection(sections, title)

	// Input fields (focusIndex 0..N-1), grouped.
	ghStyle := groupHeaderStyle(width)
	giStyle := groupInputsStyle(width)
	lastGroupID := ""
	for i := 0; i < len(m.inputMappings); i++ {
		mapping := m.inputMappings[i]

		if mapping.groupID != "" && mapping.groupID != lastGroupID {
			if lastGroupID != "" {
				sections = appendSection(sections, "\n")
			}
			groupTitle := lipgloss.JoinVertical(
				lipgloss.Top,
				groupTitleStyle.Render(mapping.groupName),
				styles.TextDim.Render(mapping.groupDescription),
			)
			sections = appendSection(sections, ghStyle.Render(groupTitle))
			lastGroupID = mapping.groupID
		}

		inputSections := []string{}

		label := labelStyle.Render(mapping.label())
		if mapping.required {
			label += styles.TextError.Render(" *")
		}
		inputSections = append(inputSections, label)

		if mapping.description != "" {
			inputSections = append(inputSections, styles.TextAccent.Render(mapping.description))
		}
		if mapping.defaultValue != "" {
			inputSections = append(inputSections, descStyle.Render(fmt.Sprintf("default: %s", mapping.defaultValue)))
		}

		fieldContent := m.inputs[i].View()
		if m.focusIndex == i {
			inputSections = append(inputSections, focusedInputStyle.Render(fieldContent))
		} else {
			inputSections = append(inputSections, blurredInputStyle.Render(fieldContent))
		}
		rendered := giStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				inputSections...,
			),
		)
		sections = appendSection(sections, rendered)
		fieldLines[i] = lineCount
	}

	// Deploy dependents toggle (focusIndex == len(inputs))
	sections = appendSection(sections, "\n")
	toggleLabel := labelStyle.Render("Deploy dependents")
	sections = appendSection(sections, toggleLabel)
	sections = appendSection(sections, descStyle.Render("Deploy components that depend on the updated inputs (press space to toggle)"))

	checkbox := "[ ] no"
	if m.deployDependents {
		checkbox = "[x] yes"
	}
	toggleDisplay := fmt.Sprintf("  %s  ", checkbox)
	if m.focusIndex == m.toggleIndex() {
		toggleDisplay = focusedInputStyle.Render(toggleDisplay)
	} else {
		toggleDisplay = blurredInputStyle.Render(toggleDisplay)
	}
	sections = appendSection(sections, toggleDisplay)
	fieldLines[m.toggleIndex()] = lineCount

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, sections...))
	m.fieldEndLines = fieldLines
}

// ensureFocusVisible scrolls the viewport so the focused field is visible.
func (m *model) ensureFocusVisible() {
	endLine, ok := m.fieldEndLines[m.focusIndex]
	if !ok {
		return
	}

	vpHeight := m.viewport.Height()
	yOffset := m.viewport.YOffset()

	if endLine > yOffset+vpHeight {
		m.viewport.SetYOffset(endLine - vpHeight)
		return
	}

	startLine := 0
	if m.focusIndex > 0 {
		if prev, ok := m.fieldEndLines[m.focusIndex-1]; ok {
			startLine = prev
		}
	}

	if startLine < yOffset {
		m.viewport.SetYOffset(startLine)
	}
}
