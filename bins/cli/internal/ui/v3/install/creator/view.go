package creator

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
// This should be called whenever the form content changes (not in View()).
func (m *model) updateViewportContent() {
	width := min(m.width, maxWidth) - 4

	// fieldLines tracks the ending line (exclusive) for each focusIndex.
	fieldLines := map[int]int{}
	lineCount := 0

	// appendSection adds a rendered section and returns how many lines it added.
	appendSection := func(sections []string, s string) []string {
		lineCount += strings.Count(s, "\n") + 1
		return append(sections, s)
	}

	sections := []string{}

	// Title
	title := titleStyle.Render("Create Install")
	if m.app != nil {
		title = titleStyle.Render(fmt.Sprintf("Create Install for %s", m.app.Name))
	}
	sections = appendSection(sections, title)

	// Name field (focusIndex 0)
	if len(m.inputMappings) > 0 {
		mapping := m.inputMappings[0]
		label := labelStyle.Render(mapping.displayName)
		if mapping.required {
			label += styles.TextError.Render(" *")
		}
		sections = appendSection(sections, label)

		if mapping.description != "" {
			sections = appendSection(sections, descStyle.Render(mapping.description))
		}

		fieldContent := m.inputs[0].View()
		if m.focusIndex == 0 {
			sections = appendSection(sections, focusedInputStyle.Render(fieldContent))
		} else {
			sections = appendSection(sections, blurredInputStyle.Render(fieldContent))
		}
		fieldLines[0] = lineCount
	}

	// Region field (focusIndex 1)
	sections = appendSection(sections, labelStyle.Render("AWS Region"))
	sections = appendSection(sections, styles.TextError.Render(" *"))
	sections = appendSection(sections, descStyle.Render("AWS region for the installation (use left/right arrows to change)"))

	regionDisplay := fmt.Sprintf("  %s  ", awsRegions[m.regionIndex])
	if m.focusIndex == 1 {
		regionDisplay = focusedInputStyle.Render(regionDisplay)
	} else {
		regionDisplay = blurredInputStyle.Render(regionDisplay)
	}
	sections = appendSection(sections, regionDisplay)
	fieldLines[1] = lineCount
	sections = appendSection(sections, "\n")

	// Dynamic input fields (focusIndex 2+), grouped
	ghStyle := groupHeaderStyle(width)
	giStyle := groupInputsStyle(width)
	lastGroupID := ""
	for i := 1; i < len(m.inputMappings); i++ {
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

		label := labelStyle.Render(mapping.displayName)
		if mapping.required {
			label += styles.TextError.Render(" *")
		}
		inputSections = append(inputSections, label)

		if mapping.description != "" {
			inputSections = append(inputSections, styles.TextAccent.Render(mapping.description))
		}

		fieldContent := m.inputs[i].View()
		if m.focusIndex == i+1 {
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
		fieldLines[i+1] = lineCount
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, sections...))
	m.fieldEndLines = fieldLines
}

// ensureFocusVisible scrolls the viewport so the focused field is visible.
// If the field's bottom is below the viewport, it scrolls down so the field
// ends at the bottom. If the field's top is above the viewport, it scrolls up.
func (m *model) ensureFocusVisible() {
	endLine, ok := m.fieldEndLines[m.focusIndex]
	if !ok {
		return
	}

	vpHeight := m.viewport.Height()
	yOffset := m.viewport.YOffset()

	// If the field ends below the visible area, scroll so it's at the bottom.
	if endLine > yOffset+vpHeight {
		m.viewport.SetYOffset(endLine - vpHeight)
		return
	}

	// Find the start line: it's the end of the previous field (or 0).
	startLine := 0
	if m.focusIndex > 0 {
		if prev, ok := m.fieldEndLines[m.focusIndex-1]; ok {
			startLine = prev
		}
	}

	// If the field starts above the visible area, scroll so it's at the top.
	if startLine < yOffset {
		m.viewport.SetYOffset(startLine)
	}
}
