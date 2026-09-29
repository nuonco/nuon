package creator

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/installcreate"
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

	helpView := m.help.View(m.keys)
	finalView.WriteString("\n")
	finalView.WriteString(lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(helpView))

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(finalView.String())
}

func (m *model) updateViewportContent() {
	if m.step == stepGroup {
		m.updateGroupViewportContent()
		return
	}

	width := min(m.width, maxWidth) - 4
	fieldLines := map[int]int{}
	lineCount := 0

	appendSection := func(sections []string, s string) []string {
		lineCount += strings.Count(s, "\n") + 1
		return append(sections, s)
	}

	sections := []string{}

	title := titleStyle.Render("Create Install")
	if m.app != nil {
		title = titleStyle.Render(fmt.Sprintf("Create Install for %s", m.app.Name))
	}
	sections = appendSection(sections, title)

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
		if m.nameChecking {
			sections = appendSection(sections, styles.TextDim.Render("Checking name availability..."))
		} else if m.nameValidationErr != nil {
			sections = appendSection(sections, warningStyle.Render(alarmIcon+m.nameValidationErr.Error()))
		}
		fieldLines[0] = lineCount
	}

	regionOffset := m.regionOffset()
	if m.needsRegion() {
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
	}

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
		if m.focusIndex == i+regionOffset {
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
		fieldLines[i+regionOffset] = lineCount
	}

	if len(m.presetLabels) > 0 {
		sections = appendSection(sections, "\n")
		sections = appendSection(sections, labelStyle.Render("Labels"))
		for k, v := range m.presetLabels {
			labelLine := fmt.Sprintf("  %s = %s",
				styles.TextAccent.Render(k),
				styles.TextDim.Render(v),
			)
			sections = appendSection(sections, labelLine)
		}
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, sections...))
	m.fieldEndLines = fieldLines
}

func (m *model) updateGroupViewportContent() {
	sections := []string{
		titleStyle.Render("Select an install group"),
		descStyle.Render("The group's labels are applied to this install so it joins the branch deployment plan."),
		"",
	}

	row := func(selected bool, label string) string {
		if selected {
			return selectedGroupStyle.Render("> " + label)
		}
		return "  " + label
	}

	for i, group := range m.groups {
		selected := i == m.groupIndex
		sections = append(sections, row(selected, group.Name))

		if labels, err := installcreate.GroupLabels(group); err == nil {
			for _, key := range slices.Sorted(maps.Keys(labels)) {
				sections = append(sections, fmt.Sprintf("      %s = %s",
					styles.TextAccent.Render(key),
					styles.TextDim.Render(labels[key]),
				))
			}
		}
	}

	skip := warningStyle.Render(alarmIcon + "Skip install group (the install will be orphaned until its labels match)")
	sections = append(sections, "", row(m.groupIndex >= len(m.groups), skip))

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Top, sections...))
	m.fieldEndLines = map[int]int{}
}

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
