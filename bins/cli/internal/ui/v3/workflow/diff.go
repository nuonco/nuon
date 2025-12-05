package workflow

import (
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/cli/styles"
	"github.com/powertoolsdev/mono/pkg/generics"

	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformDiff struct {
	FormatVersion    string
	TerraformVersion string
	OutputChanges    map[string]any
}

type HelmDiff struct {
	// wip
	Version string
}

var changeStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(styles.WarningColor)
var createStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(styles.SuccessColor)

func (m model) getTerraformDiff() string {
	var plan tfjson.Plan
	jsonBytes, err := json.Marshal(m.approvalContents.contents)
	if err != nil {
		return styles.TextError.Padding(1).Border(lipgloss.NormalBorder()).BorderForeground(styles.ErrorColor).Render(
			fmt.Sprintf("unable to marshall tf plan. \n%s", err),
		)
	}
	if err := json.Unmarshal(jsonBytes, &plan); err != nil {
		return styles.TextError.Padding(1).Border(lipgloss.NormalBorder()).BorderForeground(styles.ErrorColor).Render(
			fmt.Sprintf("unable to unmarshall tf plan. \n%s", err),
		)
	}

	// calculate updates
	updates := []*tfjson.ResourceChange{}
	creations := []*tfjson.ResourceChange{}
	deletions := []*tfjson.ResourceChange{}
	noops := []*tfjson.ResourceChange{}
	for _, rc := range plan.ResourceChanges {
		if rc.Change == nil || len(rc.Change.Actions) == 0 || generics.SliceContains(tfjson.ActionNoop, rc.Change.Actions) {
			noops = append(noops, rc)
		} else if generics.SliceContains(tfjson.ActionCreate, rc.Change.Actions) {
			creations = append(creations, rc)
		} else if generics.SliceContains(tfjson.ActionDelete, rc.Change.Actions) {
			deletions = append(deletions, rc)
		} else if generics.SliceContains(tfjson.ActionUpdate, rc.Change.Actions) {
			updates = append(updates, rc)
		}
	}

	changesSection := []string{}

	if len(creations) > 0 {
		for _, rc := range creations {
			row := createStyle.Width(m.stepDetail.Width - 4).Render(
				fmt.Sprintf("%s", rc.Address),
			)
			changesSection = append(changesSection, row)
		}
	}
	if (len(updates)) > 0 {
		for _, rc := range updates {
			row := changeStyle.Width(m.stepDetail.Width - 4).Render(
				fmt.Sprintf("%s", rc.Address),
			)
			changesSection = append(changesSection, row)
		}
	}
	if (len(updates) + len(creations)) == 0 {
		changesSection = []string{
			styles.TextSubtle.Bold(true).Margin(1, 0, 0, 0).Render("  No Changes"),
		}

	}
	return lipgloss.JoinVertical(lipgloss.Left, changesSection...)
}

func (m model) stepDetailViewStepDiff() string {
	title := styles.TextBold.Render("Resource Changes ")
	if m.approvalContents.loading {
		title = m.spinner.View() + " " + title
	}

	_, isTF := m.approvalContents.contents["terraform_version"]
	diff := ""
	if isTF {
		diff = m.getTerraformDiff()
	}

	diffSection := lipgloss.NewStyle().Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				title,
				lipgloss.NewStyle().Foreground(styles.SubtleColor).Render("[B] open in browser to see diff."),
			),
			diff,
		),
	)
	return diffSection
}
