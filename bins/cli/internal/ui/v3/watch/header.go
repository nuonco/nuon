package watch

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (m model) getOverallProgress() (int, int, float64) {
	if len(m.workflows) == 0 {
		return 0, 0, 0
	}

	totalSteps := 0
	finishedSteps := 0
	for _, w := range m.workflows {
		finished, pending, _ := common.CalculateStepProgress(w.Steps)
		totalSteps += finished + pending
		finishedSteps += finished
	}

	if totalSteps == 0 {
		return 0, 0, 0
	}

	pending := totalSteps - finishedSteps
	pct := float64(finishedSteps) / float64(totalSteps)
	return finishedSteps, pending, pct
}

func (m model) getOverallStatus() models.AppStatus {
	if len(m.workflows) == 0 {
		return models.AppStatusPending
	}

	hasInProgress := false
	hasError := false
	hasPending := false

	for _, w := range m.workflows {
		switch w.Status.Status {
		case models.AppStatusInDashProgress:
			hasInProgress = true
		case models.AppStatusError:
			hasError = true
		case models.AppStatusPending:
			hasPending = true
		}
	}

	if hasInProgress {
		return models.AppStatusInDashProgress
	}
	if hasError {
		return models.AppStatusError
	}
	if hasPending {
		return models.AppStatusPending
	}
	return models.AppStatusSuccess
}

func (m model) headerView() string {
	content := ""
	if len(m.workflows) == 0 {
		content += m.spinner.View() + " loading workflows..."
		m.header.SetContent(content)
		return appStyle.Render(m.header.View())
	}

	stepsFinished, stepsPending, progress := m.getOverallProgress()
	overallStatus := m.getOverallStatus()
	m.progress.SetPercent(progress)

	content = common.RenderHeader(common.HeaderConfig{
		Width:         m.header.Width(),
		Title:         fmt.Sprintf("Install Workflows (%d)", len(m.workflows)),
		Status:        overallStatus,
		StepsTotal:    stepsFinished + stepsPending,
		StepsFinished: stepsFinished,
		StepsPending:  stepsPending,
		Progress:      progress,
		ProgressBar:   m.progress,
		Spinner:       m.spinner,
		Actions:       m.actionsView(),
		ShowPlanOnly:  false,
		ShowSpinner:   true,
	})

	m.header.SetContent(content)
	return appStyle.Render(m.header.View())
}

func (m model) actionsView() string {
	actions := []string{}
	actions = append(actions, styles.TextDim.Render("[R] Refresh"))
	return strings.Join(actions, " ")
}
