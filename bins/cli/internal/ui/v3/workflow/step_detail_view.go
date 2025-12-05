package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
	"github.com/powertoolsdev/mono/pkg/generics"
)

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

func structToMap(obj any) (map[string]string, error) {
	// ONLY use this IFF you know all of the values are strings or can be converted to strings
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var intermediate map[string]any
	err = json.Unmarshal(jsonBytes, &intermediate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, value := range intermediate {
		switch v := value.(type) {
		case string:
			result[key] = v
		case []any:
			// Convert array to comma-separated string
			var strSlice []string
			for _, item := range v {
				strSlice = append(strSlice, fmt.Sprintf("%v", item))
			}
			result[key] = strings.Join(strSlice, ", ")
		default:
			// Convert other types to string
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

func (m model) stepDetailViewInstallStackOutputs() string {
	// NOTE(fd): aws only rn
	outputMap, err := structToMap(m.stack.InstallStackOutputs.Aws)
	if err != nil {
		return fmt.Sprintf("unable to render stack outputs\n%s", err)
	}
	//make read only table
	keys := []string{}
	maxKeyLength := 0
	for key := range outputMap {
		l := len(key)
		if l > maxKeyLength {
			maxKeyLength = l
		}
		keys = append(keys, key)
	}
	sort.Strings(keys) // sorted so order isn't all jittery
	rows := []string{}
	for i, key := range keys {
		value := outputMap[key]
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			styles.TextGhost.Render(fmt.Sprintf("[%02d] ", i))+styles.TextSubtle.Render(fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLength-len(key)))),
			" | ",
			styles.TextSubtle.Render(value), // this feels SUPER dangerous
		)
		rows = append(rows, row)
	}
	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

func (m model) stepDetailViewInstallStack() string {
	step := m.selectedStep
	s := ""
	s += lipgloss.NewStyle().Width(m.stepDetail.Width).Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.TextBold.Render("Install stack is waiting to run"),
			fmt.Sprintf("%s: %s", styles.TextSubtle.Render("Current Status"), step.Status.Status),
		),
	)

	if m.stackLoading || m.stack == nil || len(m.stack.Versions) == 0 {
		s += "\n... loading ...\n"
		return s
	}

	stack := m.stack.Versions[0]
	cliCreateCmd := fmt.Sprintf("aws cloudformation create-stack --stack-name [YOUR_STACK_NAME] --template-url %s", stack.TemplateURL)
	cliUpdateCmd := fmt.Sprintf("aws cloudformation update-stack --stack-name [YOUR_STACK_NAME] --template-url %s", stack.TemplateURL)
	s += lipgloss.NewStyle().Width(m.stepDetail.Width).Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			// install stack
			styles.TextBold.Margin(0, 0, 1).Render("Setup your install stack"),
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.TextSubtle.Render(" Install Quick Link"),
				styles.TextSubtle.Render(fmt.Sprintf(" [%s to open]", m.keys.OpenQuickLink.Help().Key)),
			),
			styles.Link.Width(m.stepDetail.Width-6).Margin(0, 1, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(stack.QuickLinkURL),
			// install template link
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styles.TextSubtle.Render(" Install Template Link"),
				styles.TextSubtle.Render(fmt.Sprintf(" [%s to open]", m.keys.OpenTemplateLink.Help().Key)),
			),
			styles.Link.Width(m.stepDetail.Width-6).Margin(0, 1, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(stack.TemplateURL),
			// divider
			styles.TextSubtle.Width(m.stepDetail.Width-6).Margin(0, 1, 1).Render(" --- or --- "),
			// CLI cmd
			styles.TextSubtle.Render(" Setup your install stack using CLI command"),
			lipgloss.NewStyle().Width(m.stepDetail.Width-6).Margin(0, 1, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(cliCreateCmd),
			// CLI update cmd
			styles.TextSubtle.Render(" Setup your install stack using CLI command"),
			lipgloss.NewStyle().Width(m.stepDetail.Width-6).Margin(0, 1, 1).Padding(1).Border(lipgloss.NormalBorder()).Render(cliUpdateCmd),
		),
	)

	if m.stack.InstallStackOutputs != nil {
		s += lipgloss.NewStyle().Width(m.stepDetail.Width).Padding(1).Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styles.TextBold.Margin(0, 0, 1).Render("Stack Outputs"),
				m.stepDetailViewInstallStackOutputs(),
			),
		)
	}
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
		content := lipgloss.NewStyle().
			Padding(1, 3).
			Render(lipgloss.JoinVertical(lipgloss.Center, "Are you sure you want to cancel this workflow?", "", "Press [C] to confirm."))
		dialog := common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.stepDetail.Width,
			Height:  m.stepDetail.Height,
			Padding: 2,
			Content: content,
			Level:   "error",
		})
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
		dialog := common.FullPageDialog(common.FullPageDialogRequest{
			Width: m.stepDetail.Width, Height: m.stepDetail.Height,
			Padding: 2,
			Content: content,
			Level:   "warning",
		})
		m.stepDetail.SetContent(dialog)
		return
	}

	// case: no selected step or a step w/ no status
	if m.selectedStep == nil || m.selectedStep.Status == nil {
		m.stepDetail.SetContent(styles.TextSubtle.Padding(3).Render("Select a workflow to get started"))
		return
	}

	sections := []string{}
	// normal case
	step := m.selectedStep

	// full-width banners
	if step.Status.Status == models.AppStatusPending {
		pendingMessage := lipgloss.NewStyle().
			Padding(2).
			Width(m.stepDetail.Width).
			Render("nothing to see at the moment...")
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
		banner := styles.SuccessBanner.
			Width(m.stepDetail.Width).
			Margin(0, 0, 1).
			Render(l1)
		sections = append(sections, banner)
	}

	// title
	style := styles.GetStatusStyle(step.Status.Status)
	title := style.
		Width(m.stepDetail.Width).
		Bold(true).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				fmt.Sprintf("%s %s", getStatusIcon(step.Status.Status), step.Name),
				styles.TextSubtle.Render(fmt.Sprintf("ID: %s", step.ID)),
			),
		)

	sections = append(sections, title)

	// stack section
	// NOTE(fd): brittle af
	if step.Name == "await install stack" || step.StepTargetType == "install_stack_versions" {
		installStack := m.stepDetailViewInstallStack()
		sections = append(sections, installStack)
	}

	// approvals section
	// TODO(fd): handle "install_sandbox_runs",
	if generics.SliceContains(step.StepTargetType, []string{"install_deploys"}) {
		diffSection := m.stepDetailViewStepDiff()
		sections = append(sections, diffSection)
	}

	jsonSection := m.stepDetailViewStepJSON()
	sections = append(sections, jsonSection)

	m.stepDetail.SetContent(lipgloss.JoinVertical(lipgloss.Left, sections...))
	if goToTop {
		m.stepDetail.GotoTop()
	}
}
