package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type HeaderConfig struct {
	Width          int
	Title          string
	Status         models.AppStatus
	StepsTotal     int
	StepsFinished  int
	StepsPending   int
	Progress       float64
	ProgressBar    progress.Model
	Spinner        spinner.Model
	Actions        string
	ShowPlanOnly   bool
	ShowSpinner    bool
	ContainerStyle lipgloss.Style
}

func RenderHeader(cfg HeaderConfig) string {
	style := styles.GetStatusStyle(cfg.Status)

	title := ""
	if cfg.ShowSpinner && cfg.Status == models.AppStatusInDashProgress {
		title += cfg.Spinner.View()
	} else {
		icon := GetStatusIcon(cfg.Status)
		title += style.Render(fmt.Sprintf("%s ", icon))
	}

	title += fmt.Sprintf("%s ", cfg.Title)
	if cfg.ShowPlanOnly {
		title += styles.TextDim.Render("[Plan Only] ")
	}
	title += fmt.Sprintf("(%02d steps)", cfg.StepsTotal)
	title += style.Render(fmt.Sprintf(" [%s]", cfg.Status))

	spacer := cfg.Spinner.View()
	spacerCount := cfg.Width - (lipgloss.Width(title) + lipgloss.Width(cfg.Actions))
	if spacerCount > 0 {
		spacer = strings.Repeat(" ", spacerCount)
	}
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, cfg.Actions)

	details := styles.TextDim.Render("  pending: ") + fmt.Sprintf("%02d ", cfg.StepsPending)
	details += styles.TextDim.Render("total:  ") + fmt.Sprintf("%02d ", cfg.StepsTotal)
	details += styles.TextDim.Render("completed: ") + fmt.Sprintf("%02d", cfg.StepsFinished)
	spacerBottom := ""
	spacerBottomCount := cfg.Width - (lipgloss.Width(details) + cfg.ProgressBar.Width)
	if spacerBottomCount > 0 {
		spacerBottom = strings.Repeat(" ", spacerBottomCount)
	}
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, details, spacerBottom, cfg.ProgressBar.ViewAs(cfg.Progress))

	content := lipgloss.JoinVertical(lipgloss.Center, topRow, bottomRow)
	return content
}
