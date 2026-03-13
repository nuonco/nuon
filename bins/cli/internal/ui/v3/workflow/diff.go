package workflow

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/generics"

	"charm.land/lipgloss/v2"
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

type terraformDiffGroups struct {
	updates   []*tfjson.ResourceChange
	creations []*tfjson.ResourceChange
	deletions []*tfjson.ResourceChange
	noops     []*tfjson.ResourceChange
}

var changeStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(styles.WarningColor)
var createStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(styles.SuccessColor)

var diffContentKeys = map[string]struct{}{
	"apply_plan_display":    {},
	"plan_display_contents": {},
	"plan_display":          {},
	"diff":                  {},
	"display":               {},
}

func collectTerraformDiffGroups(plan tfjson.Plan) terraformDiffGroups {
	groups := terraformDiffGroups{}
	classifyTerraformResourceChanges(plan.ResourceChanges, &groups)
	classifyTerraformResourceChanges(plan.ResourceDrift, &groups)
	return groups
}

func classifyTerraformResourceChanges(resourceChanges []*tfjson.ResourceChange, groups *terraformDiffGroups) {
	for _, rc := range resourceChanges {
		if rc == nil || rc.Change == nil || len(rc.Change.Actions) == 0 {
			groups.noops = append(groups.noops, rc)
			continue
		}

		actions := rc.Change.Actions
		if generics.SliceContains(tfjson.ActionNoop, actions) {
			groups.noops = append(groups.noops, rc)
			continue
		}
		if generics.SliceContains(tfjson.ActionCreate, actions) {
			groups.creations = append(groups.creations, rc)
			continue
		}
		if generics.SliceContains(tfjson.ActionDelete, actions) {
			groups.deletions = append(groups.deletions, rc)
			continue
		}
		if generics.SliceContains(tfjson.ActionUpdate, actions) {
			groups.updates = append(groups.updates, rc)
			continue
		}

		groups.noops = append(groups.noops, rc)
	}
}

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

	groups := collectTerraformDiffGroups(plan)

	changesSection := []string{}

	if len(groups.creations) > 0 {
		for _, rc := range groups.creations {
			row := createStyle.Width(m.stepDetail.Width() - 4).Render(
				rc.Address,
			)
			changesSection = append(changesSection, row)
		}
	}
	if len(groups.updates) > 0 {
		for _, rc := range groups.updates {
			row := changeStyle.Width(m.stepDetail.Width() - 4).Render(
				rc.Address,
			)
			changesSection = append(changesSection, row)
		}
	}
	if len(groups.deletions) > 0 {
		for _, rc := range groups.deletions {
			row := styles.TextError.Border(lipgloss.NormalBorder()).BorderForeground(styles.ErrorColor).
				Width(m.stepDetail.Width() - 4).
				Render(rc.Address)
			changesSection = append(changesSection, row)
		}
	}
	if len(groups.updates)+len(groups.creations)+len(groups.deletions) == 0 {
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
	if m.approvalContents.error != nil {
		errBlock := styles.TextError.Padding(1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.ErrorColor).
			Render(fmt.Sprintf("unable to load diff contents:\n%s", m.approvalContents.error))

		return lipgloss.NewStyle().Padding(1).Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					title,
					lipgloss.NewStyle().Foreground(styles.SubtleColor).Render("[B] open in browser for full context."),
				),
				errBlock,
			),
		)
	}

	_, isTF := m.approvalContents.contents["terraform_version"]
	sections := []string{}
	if isTF {
		sections = append(sections, m.getTerraformDiff())
	}

	if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
		sections = append(sections, m.renderDiffText(fullDiff))
	}

	if len(sections) == 0 {
		sections = append(sections, m.renderRawDiffFallback())
	}

	diffSection := lipgloss.NewStyle().Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				title,
				lipgloss.NewStyle().Foreground(styles.SubtleColor).Render("[B] open in browser for full context."),
			),
			lipgloss.JoinVertical(lipgloss.Left, sections...),
		),
	)
	return diffSection
}

func (m model) renderDiffText(text string) string {
	content := strings.TrimSpace(text)
	if content == "" {
		content = "No diff contents available"
	}

	width := m.stepDetail.Width() - 4
	if width < 20 {
		width = 20
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(width).
		Render(content)
}

func (m model) renderRawDiffFallback() string {
	if m.approvalContents.raw == nil {
		return styles.TextSubtle.Padding(1).Render("No diff contents available")
	}

	formatted, err := json.MarshalIndent(m.approvalContents.raw, "", "  ")
	if err != nil {
		return styles.TextSubtle.Padding(1).Render(fmt.Sprintf("No renderable diff contents (%v)", err))
	}

	return m.renderDiffText(string(formatted))
}

func extractDisplayDiffText(value any) string {
	if value == nil {
		return ""
	}

	if text := findDisplayDiffText(value); text != "" {
		return text
	}

	if direct, ok := value.(string); ok {
		return strings.TrimSpace(direct)
	}

	return ""
}

func findDisplayDiffText(value any) string {
	switch v := value.(type) {
	case map[string]any:
		for k, raw := range v {
			if _, ok := diffContentKeys[strings.ToLower(k)]; ok {
				if text := coerceDisplayDiffText(raw); text != "" {
					return text
				}
			}
		}

		for _, raw := range v {
			if text := findDisplayDiffText(raw); text != "" {
				return text
			}
		}
	case []any:
		for _, raw := range v {
			if text := findDisplayDiffText(raw); text != "" {
				return text
			}
		}
	}

	return ""
}

func coerceDisplayDiffText(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	case []int64:
		bytes := make([]byte, 0, len(v))
		for _, n := range v {
			if n < 0 || n > 255 {
				return ""
			}
			bytes = append(bytes, byte(n))
		}
		return strings.TrimSpace(string(bytes))
	case []any:
		bytes := make([]byte, 0, len(v))
		for _, n := range v {
			num, ok := n.(float64)
			if !ok || num < 0 || num > 255 || num != math.Trunc(num) {
				return ""
			}
			bytes = append(bytes, byte(num))
		}
		return strings.TrimSpace(string(bytes))
	}

	return ""
}
