package workflow

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var dialogBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(1, 0).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)

func (m model) stepIsApprovable() bool {
	if m.selectedStep == nil {
		return false
	}

	step := m.selectedStep
	if step.Approval == nil {
		return false
	} else if step.Approval.Response == nil {
		return true
	}
	return false

}

func (m model) StepDetailApprovalRequiredBanner() string {
	s := styles.ApprovalConfirmation.
		Width(m.stepDetail.Width).Margin(0, 0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Center,
			"approval required",
			"",
			"press [\"a\"] to approve this step.",
		))
	return s
}

func (m model) stepDetailViewApprovalConfirmationBanner() string {
	s := lipgloss.JoinVertical(
		lipgloss.Center,
		"Are you sure you want to approve this step?",
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Padding(0, 2).Render("[a] Approve"),
			" ",
			lipgloss.NewStyle().Padding(0, 2).Render("[esc] Cancel"),
		),
	)
	return styles.ApprovalConfirmation.Width(m.stepDetail.Width).Margin(0, 0, 1).Render(s)
}

func (m model) stepDetailViewStepJSON() string {
	// takes the m.selectedStep and renders it as a indented string
	if m.selectedStep == nil {
		return ""
	}
	caret := "▾"
	if !m.showJson {
		caret = "▸"
	}
	header := styles.TextBold.Padding(0, 1).Render(fmt.Sprintf("%s Step JSON", caret))
	if !m.showJson {
		return header
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).Margin(0, 1).
		Width(m.stepDetail.Width - 4)
	jsonBytes, err := json.MarshalIndent(m.selectedStep, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling step JSON: %v", err)
	}
	body := style.Render(string(jsonBytes))
	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		"",
		body,
	)
}

func (m model) stepDetailViewInstallStack() string {
	step := m.selectedStep
	s := ""
	s += lipgloss.NewStyle().Width(m.stepDetail.Width).Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.TextBold.Render("Install stack is waiting to run"),
			fmt.Sprintf("%s: %s", styles.TextDim.Render("Current Status"), step.Status.Status),
		),
	)

	if m.stackLoading || m.stack == nil || len(m.stack.Versions) == 0 {
		s += "\n... loading ...\n"
		return s
	}

	stack := m.stack.Versions[0]
	s += lipgloss.NewStyle().Width(m.stepDetail.Width).Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			// install stack
			styles.TextBold.Render("Setup your install stack"),
			"",
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.TextLight.Render(" Install Quick Link"),
				styles.TextDim.Render(fmt.Sprintf(" [%s to open]", m.keys.OpenQuickLink.Help().Key)),
			),
			styles.Link.Width(m.stepDetail.Width-6).Margin(0, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(stack.QuickLinkURL),
			"",
			// install template link
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.TextLight.Render(" Install Template Link"),
				styles.TextDim.Render(fmt.Sprintf(" [%s to open]", m.keys.OpenTemplateLink.Help().Key)),
			),
			styles.Link.Width(m.stepDetail.Width-6).Margin(0, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(stack.TemplateURL),
		),
	)
	return s
}

func (m *model) populateStepDetailView(goToTop bool) {
	// loading states: exit early
	if len(m.steps) == 0 {
		s := "\n\n\tLoading ...\n"
		m.stepDetail.SetContent(s)
		return
	}

	// case: workflow cancellation confirmation prompt
	if m.workflowCancelationConf {
		// in this case, we hijack the view to show a big red confirmation
		content := lipgloss.NewStyle().Padding(1, 3).Render(lipgloss.JoinVertical(lipgloss.Center, "Are you sure you want to cancel this workflow?", "", "Press [C] to confirm."))
		dialog := common.FullPageDialog(common.FullPageDialogRequest{Width: m.stepDetail.Width, Height: m.stepDetail.Height, Padding: 2, Content: content, Level: "warning"})
		m.stepDetail.SetContent(dialog)
		return
	}

	// case: workflow approve all confirmation prompt
	if m.workflowApprovalConf {
		// in this case, we hijack the view to show a big confirmation
		content := lipgloss.NewStyle().Padding(1, 3).
			Render(
				lipgloss.JoinVertical(lipgloss.Center, "Are you sure you want to approve all?", "", "Press [A] to confirm."),
			)
		dialog := common.FullPageDialog(common.FullPageDialogRequest{Width: m.stepDetail.Width, Height: m.stepDetail.Height, Padding: 2, Content: content, Level: "warning"})
		m.stepDetail.SetContent(dialog)
		return
	}

	// case: no selected step or a step w/ no status
	if m.selectedStep == nil || m.selectedStep.Status == nil {
		m.stepDetail.SetContent(styles.TextDim.Padding(3).Render("Select a workflow to get started"))
		return
	}

	sections := []string{}
	// normal case
	step := m.selectedStep
	style := styles.GetStatusStyle(step.Status.Status)
	title := style.
		Width(m.stepDetail.Width).
		Bold(true).
		Padding(1).
		Render(fmt.Sprintf("%s %s", getStatusIcon(step.Status.Status), step.Name))

	sections = append(sections, title)

	if step.Status.Status == models.AppStatusPending {
		pendingMessage := lipgloss.NewStyle().Padding(2).Width(m.stepDetail.Width).Render("nothing to see at the moment...")
		sections = append(sections, pendingMessage)
	}

	if step.Status.Status == models.AppStatusApprovalDashAwaiting {
		if m.stepApprovalConf {
			banner := m.stepDetailViewApprovalConfirmationBanner()
			sections = append(sections, banner)
		} else {
			banner := m.StepDetailApprovalRequiredBanner() + "\n"
			sections = append(sections, banner)
		}
	} else if step.Approval != nil && step.Approval.Response != nil {
		// TODO: make this a re-usable component
		l1 := lipgloss.NewStyle().Bold(true).Render("Plan Approved") + "\n"
		l1 += "These changes have been approved and changes will be applied."
		banner := styles.SuccessBanner.Width(m.stepDetail.Width).Margin(0, 0, 1).Render(l1)
		sections = append(sections, banner)
	}

	// NOTE(fd): brittle af
	if step.Name == "await install stack" || step.StepTargetType == "install_stack_versions" {
		installStack := m.stepDetailViewInstallStack()
		sections = append(sections, installStack)
	}

	jsonSection := m.stepDetailViewStepJSON()
	sections = append(sections, jsonSection)

	m.stepDetail.SetContent(lipgloss.JoinVertical(lipgloss.Left, sections...))
	if goToTop {
		m.stepDetail.GotoTop()
	}
}
