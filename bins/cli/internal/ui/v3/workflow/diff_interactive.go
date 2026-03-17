package workflow

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

type helmDiffExplorerModel struct {
	width         int
	stepID        string
	plan          parsedHelmPlan
	hasPlan       bool
	parseErr      error
	selectedIndex int
	expanded      map[int]bool
}

func newHelmDiffExplorerModel(width int) helmDiffExplorerModel {
	m := helmDiffExplorerModel{
		expanded: map[int]bool{},
	}
	m.SetWidth(width)
	return m
}

func (m *helmDiffExplorerModel) SetWidth(width int) {
	if width < 20 {
		width = 20
	}

	m.width = width
}

func (m *helmDiffExplorerModel) Reset() {
	m.stepID = ""
	m.plan = parsedHelmPlan{}
	m.hasPlan = false
	m.parseErr = nil
	m.selectedIndex = 0
	m.expanded = map[int]bool{}
}

func (m *helmDiffExplorerModel) Bind(stepID string, raw any, contents map[string]interface{}) {
	if m.expanded == nil {
		m.expanded = map[int]bool{}
	}

	stepChanged := m.stepID != stepID
	m.stepID = stepID

	if stepChanged {
		m.selectedIndex = 0
		m.expanded = map[int]bool{}
	}

	if raw == nil && len(contents) == 0 {
		m.plan = parsedHelmPlan{}
		m.hasPlan = false
		m.parseErr = nil
		return
	}

	plan, err := parseHelmPlan(raw)
	if err != nil {
		if fallbackPlan, fallbackErr := parseHelmPlan(contents); fallbackErr == nil {
			plan = fallbackPlan
			err = nil
		}
	}

	if err != nil {
		m.plan = parsedHelmPlan{}
		m.hasPlan = false
		m.parseErr = err
		return
	}

	m.plan = plan
	m.hasPlan = true
	m.parseErr = nil

	if len(m.plan.changes) == 0 {
		m.selectedIndex = 0
		m.expanded = map[int]bool{}
		return
	}

	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
	if m.selectedIndex >= len(m.plan.changes) {
		m.selectedIndex = len(m.plan.changes) - 1
	}

	for idx := range m.expanded {
		if idx < 0 || idx >= len(m.plan.changes) {
			delete(m.expanded, idx)
		}
	}
}

func (m *helmDiffExplorerModel) Update(msg tea.KeyPressMsg) bool {
	if !m.hasInteractiveRows() {
		return false
	}

	switch msg.String() {
	case "k":
		m.moveSelection(-1)
		return true
	case "j":
		m.moveSelection(1)
		return true
	case "enter":
		return m.toggleSelectedExpanded()
	}

	if msg.Code == tea.KeySpace || msg.Text == " " {
		return m.toggleSelectedExpanded()
	}

	return false
}

func (m *helmDiffExplorerModel) moveSelection(delta int) bool {
	if !m.hasInteractiveRows() {
		return false
	}

	next := m.selectedIndex + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.plan.changes) {
		next = len(m.plan.changes) - 1
	}

	if next == m.selectedIndex {
		return false
	}

	m.selectedIndex = next
	return true
}

func (m *helmDiffExplorerModel) toggleSelectedExpanded() bool {
	if !m.hasInteractiveRows() {
		return false
	}

	if m.expanded[m.selectedIndex] {
		delete(m.expanded, m.selectedIndex)
		return true
	}

	m.expanded = map[int]bool{m.selectedIndex: true}
	return true
}

func (m helmDiffExplorerModel) hasInteractiveRows() bool {
	return m.hasPlan && len(m.plan.changes) > 0
}

func (m helmDiffExplorerModel) View() string {
	if m.parseErr != nil {
		return styles.TextError.Padding(1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.ErrorColor).
			Render(fmt.Sprintf("unable to parse diff contents:\n%s", m.parseErr))
	}

	if !m.hasPlan {
		return styles.TextSubtle.Padding(1).Render("No diff contents available")
	}

	plan := m.plan
	overviewTitle := "Helm changes overview"
	if plan.planSource == "kubernetes" {
		overviewTitle = "Kubernetes changes overview"
	}

	if plan.summary.add == 0 && plan.summary.change == 0 && plan.summary.destroy == 0 {
		for _, change := range plan.changes {
			incrementHelmSummaryByAction(&plan.summary, change.action)
		}
	}

	sections := []string{
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.BorderInactiveColor).
			Width(m.width).
			Padding(1).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					styles.TextBold.Render(overviewTitle),
					styles.TextSubtle.Render(fmt.Sprintf("Operation: %s", strings.TrimSpace(plan.op))),
				),
			),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.BorderInactiveColor).
			Width(m.width).
			Padding(0, 1).
			Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					styles.TextSuccess.Bold(true).Render(fmt.Sprintf("%d", plan.summary.add)),
					" ",
					styles.TextSubtle.Render("to add"),
					"   ",
					styles.TextWarning.Bold(true).Render(fmt.Sprintf("%d", plan.summary.change)),
					" ",
					styles.TextSubtle.Render("to change"),
					"   ",
					styles.TextError.Bold(true).Render(fmt.Sprintf("%d", plan.summary.destroy)),
					" ",
					styles.TextSubtle.Render("to destroy"),
				),
			),
	}

	if len(plan.changes) == 0 {
		sections = append(
			sections,
			lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.BorderInactiveColor).
				Width(m.width).
				Padding(1).
				Render(styles.TextSubtle.Render("No changes found.")),
		)

		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	sections = append(
		sections,
		styles.TextSubtle.Margin(0, 1).Render("Use [j/k] to select, [enter] to expand, and [↑/↓] to scroll."),
	)

	for i, change := range plan.changes {
		diff := findHelmDiffForChange(change, plan.content)

		diffText := ""
		if diff != nil {
			diffText = renderHelmDiffText(*diff)
		}

		actionLabel := normalizeHelmAction(change.action)
		actionColor := styles.WarningColor
		borderColor := styles.WarningColor
		switch actionLabel {
		case "added":
			actionColor = styles.SuccessColor
			borderColor = styles.SuccessColor
		case "destroyed":
			actionColor = styles.ErrorColor
			borderColor = styles.ErrorColor
		}

		isSelected := i == m.selectedIndex
		isExpanded := m.expanded[i]

		indicator := "▸"
		if isExpanded {
			indicator = "▾"
		}

		releaseHeader := styles.TextBold.Render(fmt.Sprintf("%s %s", indicator, change.release))
		actionHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(actionColor).
			Render(actionLabel)

		headers := []string{
			renderDiffHeaderRow(releaseHeader, actionHeader, m.width-4),
			styles.TextSubtle.Render(fmt.Sprintf("%s (%s)", change.resource, change.resourceType)),
			styles.TextSubtle.Render(fmt.Sprintf("Namespace: %s", change.namespace)),
		}

		if isSelected && !isExpanded {
			headers = append(headers, styles.TextSubtle.Render("Press [enter] to expand this diff."))
		}

		if isExpanded {
			if strings.TrimSpace(diffText) == "" {
				headers = append(
					headers,
					lipgloss.NewStyle().
						MarginTop(1).
						Padding(1).
						Border(lipgloss.NormalBorder()).
						BorderForeground(styles.BorderInactiveColor).
						Render(styles.TextSubtle.Render("No diff available for this change.")),
				)
			} else {
				headers = append(headers, renderHelmDiffTextWithWidth(diffText, m.width-4))
			}
		}

		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Width(m.width).
			Padding(1)
		if isSelected {
			cardStyle = cardStyle.BorderForeground(styles.AccentColor)
		} else {
			cardStyle = cardStyle.BorderForeground(borderColor)
		}

		sections = append(sections, cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, headers...)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) syncHelmDiffExplorer() {
	m.helmDiffExplorer.SetWidth(m.stepDetail.Width() - 4)

	if m.selectedStep == nil || !m.stepHasPlanDiff(m.selectedStep) {
		return
	}

	if !isInteractivePlanDiffType(m.selectedStepDiffType()) {
		return
	}

	if m.approvalContents.loading {
		return
	}

	m.helmDiffExplorer.Bind(
		m.selectedStep.ID,
		m.approvalContents.raw,
		m.approvalContents.contents,
	)
}

func (m *model) handleDetailContentKey(msg tea.KeyPressMsg) bool {
	if m.focus != "detail" || m.selectedStep == nil {
		return false
	}

	if m.workflowCancelationConf || m.workflowApprovalConf || m.stepApprovalConf {
		return false
	}

	if !m.stepHasPlanDiff(m.selectedStep) || !isInteractivePlanDiffType(m.selectedStepDiffType()) {
		return false
	}

	if msg.String() == "up" || msg.String() == "down" {
		return false
	}

	m.syncHelmDiffExplorer()

	return m.helmDiffExplorer.Update(msg)
}
