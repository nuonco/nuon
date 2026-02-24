package watch

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (m *model) populateWorkflowDetailView() {
	if m.selectedWorkflow == nil {
		m.workflowDetail.SetContent(styles.TextDim.Render("Select a workflow to view details"))
		return
	}

	w := m.selectedWorkflow
	content := m.renderWorkflowDetail(w)
	m.workflowDetail.SetContent(content)
}

func (m *model) renderWorkflowDetail(w *models.AppWorkflow) string {
	var sb strings.Builder

	icon := common.GetStatusIcon(w.Status.Status)
	statusStyle := styles.GetStatusStyle(w.Status.Status)

	sb.WriteString(statusStyle.Bold(true).Render(fmt.Sprintf("%s %s", icon, w.Name)))
	sb.WriteString("\n\n")

	sb.WriteString(styles.TextDim.Render("ID: "))
	sb.WriteString(w.ID)
	sb.WriteString("\n")

	sb.WriteString(styles.TextDim.Render("Status: "))
	sb.WriteString(statusStyle.Render(string(w.Status.Status)))
	sb.WriteString("\n")

	if w.PlanOnly {
		sb.WriteString(styles.TextDim.Render("Mode: "))
		sb.WriteString(lipgloss.NewStyle().Foreground(styles.WarningColor).Render("Plan Only"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	finished, pending, pct := common.CalculateStepProgress(w.Steps)
	sb.WriteString(styles.TextDim.Render("Progress: "))
	sb.WriteString(fmt.Sprintf("%.0f%% (%d/%d steps)\n", pct*100, finished, finished+pending))

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Steps:"))
	sb.WriteString("\n")

	for _, step := range w.Steps {
		stepIcon := common.GetStatusIcon(step.Status.Status)
		stepStyle := styles.GetStatusStyle(step.Status.Status)
		stepLine := fmt.Sprintf("  %s [%02d] %s",
			stepStyle.Render(stepIcon),
			step.Idx,
			step.Name,
		)
		sb.WriteString(stepLine)
		sb.WriteString(" ")
		sb.WriteString(stepStyle.Render(string(step.Status.Status)))
		if step.Finished {
			sb.WriteString(styles.TextDim.Render(fmt.Sprintf(" (%s)", common.HumanizeNSDuration(step.ExecutionTime))))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
