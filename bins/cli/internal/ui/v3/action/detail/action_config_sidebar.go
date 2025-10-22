package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m *Model) populateActionConfigView(setContent bool) {
	if !setContent {
		return
	}

	content := []string{}

	// If no config loaded yet
	if m.latestConfig == nil {
		if m.configLoading {
			content = append(content, "Loading latest configuration...")
		} else {
			content = append(content, "No configuration available")
		}
		m.actionConfig.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
		return
	}

	// If no steps in config
	if len(m.latestConfig.Steps) == 0 {
		content = append(content, "No steps configured")
		m.actionConfig.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
		return
	}

	// Render header
	content = append(content, m.renderStepsHeader())
	// Render each step
	for i, step := range m.latestConfig.Steps {
		content = append(content, m.renderStep(step, i+1))
	}

	m.actionConfig.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
}

func (m Model) renderStepsHeader() string {
	return styles.TextBold.Render("Latest Configured Steps")
}

func (m Model) renderStep(step *models.AppActionWorkflowStepConfig, stepNumber int) string {
	if step == nil {
		return ""
	}

	sections := []string{}

	// Step header
	stepHeader := styles.TextSubtle.Render(fmt.Sprintf("Step %d", stepNumber))
	if step.Name != "" {
		stepHeader += ": " + step.Name
	}
	sections = append(sections, stepHeader)

	// Repository Details section
	repoSection := m.renderRepositoryDetails(step)
	if repoSection != "" {
		sections = append(sections, "", m.renderSectionHeader("Repository Details"), repoSection)
	}

	// Command section
	commandSection := m.renderCommand(step)
	if commandSection != "" {
		sections = append(sections, "", m.renderSectionHeader("Command"), commandSection)
	}

	// Variables section
	varsSection := m.renderVariables(step)
	if varsSection != "" {
		sections = append(sections, "", m.renderSectionHeader("Variables"), varsSection)
	}

	// Join all sections with a border
	stepContent := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Dim).
		Padding(1).
		Width(m.stepsWidth - 6).
		Render(stepContent)
}

func (m Model) renderSectionHeader(title string) string {
	return styles.TextDim.Underline(true).Render(title)
}

func (m Model) renderRepositoryDetails(step *models.AppActionWorkflowStepConfig) string {
	details := []string{}

	// Connected GitHub VCS
	if step.ConnectedGithubVcsConfig != nil {
		if step.ConnectedGithubVcsConfig.Repo != "" {
			details = append(details, fmt.Sprintf("repo: %s", step.ConnectedGithubVcsConfig.Repo))
		}
		if step.ConnectedGithubVcsConfig.Branch != "" {
			details = append(details, fmt.Sprintf("branch: %s", step.ConnectedGithubVcsConfig.Branch))
		}
		if step.ConnectedGithubVcsConfig.Directory != "" {
			details = append(details, fmt.Sprintf("directory: %s", step.ConnectedGithubVcsConfig.Directory))
		}
	}

	// Public Git VCS
	if step.PublicGitVcsConfig.Repo != "" {
		details = append(details, fmt.Sprintf("repo: %s", step.PublicGitVcsConfig.Repo))
	}
	if step.PublicGitVcsConfig.Branch != "" {
		details = append(details, fmt.Sprintf("branch: %s", step.PublicGitVcsConfig.Branch))
	}
	if step.PublicGitVcsConfig.Directory != "" {
		details = append(details, fmt.Sprintf("directory: %s", step.PublicGitVcsConfig.Directory))
	}

	if len(details) == 0 {
		return ""
	}

	return strings.Join(details, "\n")
}

func (m Model) renderCommand(step *models.AppActionWorkflowStepConfig) string {
	if step.Command == "" && step.InlineContents == "" {
		return ""
	}

	codeStyle := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Background(lipgloss.Color("235")).
		Padding(1).
		Width(m.stepsWidth - 8)

	if step.Command != "" {
		return codeStyle.Render(step.Command)
	}
	if step.InlineContents != "" {
		return codeStyle.Render(step.InlineContents)
	}

	return ""
}

func (m Model) renderVariables(step *models.AppActionWorkflowStepConfig) string {
	if len(step.EnvVars) == 0 {
		return styles.TextBold.BorderBottom(true).BorderForeground(styles.SubtleColor).Render("no env vars")
	}

	rows := []string{}
	for key, value := range step.EnvVars {
		// Truncate long values
		rows = append(rows, fmt.Sprintf("%s:\n  %s", styles.TextDim.Width(m.stepsWidth-4).Render(key), value))
	}

	return strings.Join(rows, "\n")
}
