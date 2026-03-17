package workflow

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

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

type helmPlanSummary struct {
	add     int
	change  int
	destroy int
}

type helmPlanChange struct {
	namespace    string
	release      string
	resource     string
	resourceType string
	action       string
}

type helmContentDiffEntry struct {
	Delta   any            `json:"delta"`
	Type    any            `json:"type"`
	Path    flexibleString `json:"path"`
	Payload flexibleString `json:"payload"`
}

type helmContentDiff struct {
	Version   any                    `json:"_version"`
	API       flexibleString         `json:"api"`
	Name      flexibleString         `json:"name"`
	Namespace flexibleString         `json:"namespace"`
	Kind      flexibleString         `json:"kind"`
	Resource  flexibleString         `json:"resource"`
	Op        flexibleString         `json:"op"`
	Type      flexibleString         `json:"type"`
	DryRun    any                    `json:"dry_run"`
	Before    flexibleString         `json:"before"`
	After     flexibleString         `json:"after"`
	Entries   []helmContentDiffEntry `json:"entries"`
}

type helmPlanPayload struct {
	Op              flexibleString    `json:"op"`
	Plan            flexibleString    `json:"plan"`
	HelmContentDiff []helmContentDiff `json:"helm_content_diff"`
	K8sContentDiff  []helmContentDiff `json:"k8s_content_diff"`
	DryRunOutput    flexibleString    `json:"dry_run_output"`
}

type parsedHelmPlan struct {
	op         string
	changes    []helmPlanChange
	summary    helmPlanSummary
	content    []helmContentDiff
	planSource string
	dryRun     string
}

type componentDiffType string

const (
	componentDiffTypeUnknown            componentDiffType = "unknown"
	componentDiffTypeTerraform          componentDiffType = "terraform"
	componentDiffTypeHelm               componentDiffType = "helm"
	componentDiffTypeKubernetesManifest componentDiffType = "kubernetes_manifest"

	helmDiffContextLines      = 1
	helmDiffCollapseMinLines  = 12
	helmDiffCollapseMinHidden = 3
)

type helmRenderedDiffLine struct {
	raw         string
	indent      int
	isChanged   bool
	isContainer bool
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*s = flexibleString(coerceAnyToString(value))
	return nil
}

func coerceAnyToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []any, map[string]any:
		if text := coerceDisplayDiffText(v); text != "" {
			return text
		}

		bytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}

		return string(bytes)
	default:
		return fmt.Sprintf("%v", v)
	}
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

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)
var helmPlanLineRe = regexp.MustCompile(`^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)\s*to\s*be\s*(\w+)`)
var helmSummaryRe = regexp.MustCompile(`Plan:\s*(\d+)\s+to\s+add,\s*(\d+)\s+to\s+change,\s*(\d+)\s+to\s+destroy`)

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

func (m *model) stepDetailViewStepDiff() string {
	title := styles.TextBold.Render("Resource Changes ")
	if m.approvalContents.loading {
		title = m.spinner.View() + " " + title
	}

	hint := "[B] open in browser for full context."
	if isInteractivePlanDiffType(m.selectedStepDiffType()) && m.helmDiffExplorer.hasInteractiveRows() {
		hint = "[j/k] select [enter] expand [↑/↓] scroll [B] browser"
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
					lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(hint),
				),
				errBlock,
			),
		)
	}

	sections := []string{}
	switch m.selectedStepDiffType() {
	case componentDiffTypeTerraform:
		sections = append(sections, m.stepDetailViewTerraformDiff())
	case componentDiffTypeHelm:
		sections = append(sections, m.stepDetailViewHelmDiff())
	case componentDiffTypeKubernetesManifest:
		sections = append(sections, m.stepDetailViewKubernetesManifestDiff())
	default:
		if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
			sections = append(sections, m.renderDiffText(fullDiff))
		}
		if len(sections) == 0 {
			sections = append(sections, m.renderRawDiffFallback())
		}
	}

	diffSection := lipgloss.NewStyle().Padding(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				title,
				lipgloss.NewStyle().Foreground(styles.SubtleColor).Render(hint),
			),
			lipgloss.JoinVertical(lipgloss.Left, sections...),
		),
	)
	return diffSection
}

func (m model) selectedStepDiffType() componentDiffType {
	if m.selectedStep != nil && m.selectedStep.Approval != nil {
		switch m.selectedStep.Approval.Type {
		case models.AppWorkflowStepApprovalTypeTerraformPlan:
			return componentDiffTypeTerraform
		case models.AppWorkflowStepApprovalTypeHelmApproval:
			return componentDiffTypeHelm
		case models.AppWorkflowStepApprovalTypeKubernetesManifestApproval:
			return componentDiffTypeKubernetesManifest
		}
	}

	if _, isTerraform := m.approvalContents.contents["terraform_version"]; isTerraform {
		return componentDiffTypeTerraform
	}
	if _, isHelm := m.approvalContents.contents["helm_content_diff"]; isHelm {
		return componentDiffTypeHelm
	}
	if _, isKubernetesManifest := m.approvalContents.contents["k8s_content_diff"]; isKubernetesManifest {
		return componentDiffTypeKubernetesManifest
	}

	return componentDiffTypeUnknown
}

func (m model) stepDetailViewTerraformDiff() string {
	sections := []string{m.getTerraformDiff()}

	if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
		sections = append(sections, m.renderDiffText(fullDiff))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) stepDetailViewHelmDiff() string {
	m.syncHelmDiffExplorer()

	if m.approvalContents.loading && !m.helmDiffExplorer.hasPlan {
		return styles.TextSubtle.Padding(1).Render(m.spinner.View() + " loading helm diff contents...")
	}

	if m.helmDiffExplorer.parseErr != nil {
		sections := []string{
			styles.TextError.Padding(1).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.ErrorColor).
				Render(fmt.Sprintf("unable to parse helm diff contents:\n%s", m.helmDiffExplorer.parseErr)),
		}

		if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
			sections = append(sections, m.renderDiffText(fullDiff))
		} else {
			sections = append(sections, m.renderRawDiffFallback())
		}

		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	if m.helmDiffExplorer.hasPlan {
		return m.helmDiffExplorer.View()
	}

	if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
		return m.renderDiffText(fullDiff)
	}

	return m.renderRawDiffFallback()
}

func (m *model) stepDetailViewKubernetesManifestDiff() string {
	m.syncHelmDiffExplorer()

	if m.approvalContents.loading && !m.helmDiffExplorer.hasPlan {
		return styles.TextSubtle.Padding(1).Render(m.spinner.View() + " loading kubernetes manifest diff contents...")
	}

	sections := []string{}

	if m.helmDiffExplorer.parseErr != nil {
		sections = append(
			sections,
			styles.TextError.Padding(1).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.ErrorColor).
				Render(fmt.Sprintf("unable to parse kubernetes manifest diff contents:\n%s", m.helmDiffExplorer.parseErr)),
		)
	} else if m.helmDiffExplorer.hasPlan {
		sections = append(sections, m.helmDiffExplorer.View())
	}

	if len(sections) == 0 {
		if fullDiff := extractDisplayDiffText(m.approvalContents.raw); fullDiff != "" {
			sections = append(sections, m.renderDiffText(fullDiff))
		}
	}

	if len(sections) == 0 || !m.helmDiffExplorer.hasInteractiveRows() {
		if dryRunOutput := extractDryRunOutput(m.approvalContents.raw); dryRunOutput != "" {
			sections = append(sections, renderManifestContextSection("Dry-run manifest output", m.renderDiffTextWithWidth(dryRunOutput, m.stepDetail.Width()-4)))
		}
	}

	if len(sections) == 0 {
		sections = append(sections, m.renderRawDiffFallback())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func parseHelmPlan(value any) (parsedHelmPlan, error) {
	if value == nil {
		return parsedHelmPlan{}, fmt.Errorf("empty helm payload")
	}

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return parsedHelmPlan{}, fmt.Errorf("unable to marshal helm payload: %w", err)
	}

	var payload helmPlanPayload
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return parsedHelmPlan{}, fmt.Errorf("unable to unmarshal helm payload: %w", err)
	}

	hydrateNestedPlanPayload(&payload)

	planSource := "helm"
	content := payload.HelmContentDiff
	if len(content) == 0 && len(payload.K8sContentDiff) > 0 {
		planSource = "kubernetes"
		content = payload.K8sContentDiff
	}

	changes, summary := parseHelmPlanText(string(payload.Plan))
	if len(changes) == 0 {
		changes = synthesizeHelmChanges(content)
	}

	return parsedHelmPlan{
		op:         strings.TrimSpace(string(payload.Op)),
		changes:    changes,
		summary:    summary,
		content:    content,
		planSource: planSource,
		dryRun:     strings.TrimSpace(string(payload.DryRunOutput)),
	}, nil
}

func hydrateNestedPlanPayload(payload *helmPlanPayload) {
	if payload == nil {
		return
	}

	nested, ok := parseHelmPlanPayloadString(string(payload.Plan))
	if !ok {
		return
	}

	if strings.TrimSpace(string(payload.Op)) == "" {
		payload.Op = nested.Op
	}
	if len(payload.HelmContentDiff) == 0 && len(nested.HelmContentDiff) > 0 {
		payload.HelmContentDiff = nested.HelmContentDiff
	}
	if len(payload.K8sContentDiff) == 0 && len(nested.K8sContentDiff) > 0 {
		payload.K8sContentDiff = nested.K8sContentDiff
	}
	if strings.TrimSpace(string(payload.DryRunOutput)) == "" {
		payload.DryRunOutput = nested.DryRunOutput
	}

	nestedPlan := strings.TrimSpace(string(nested.Plan))
	if nestedPlan != "" && !looksLikeJSONPayload(nestedPlan) {
		payload.Plan = flexibleString(nestedPlan)
	}
}

func parseHelmPlanPayloadString(plan string) (helmPlanPayload, bool) {
	trimmed := strings.TrimSpace(plan)
	if !looksLikeJSONPayload(trimmed) {
		return helmPlanPayload{}, false
	}

	var payload helmPlanPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return helmPlanPayload{}, false
	}

	return payload, true
}

func looksLikeJSONPayload(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func parseHelmPlanText(planText string) ([]helmPlanChange, helmPlanSummary) {
	clean := ansiEscapeRe.ReplaceAllString(planText, "")
	lines := strings.Split(clean, "\n")
	changes := []helmPlanChange{}
	summary := helmPlanSummary{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if matches := helmPlanLineRe.FindStringSubmatch(trimmed); len(matches) == 6 {
			changes = append(changes, helmPlanChange{
				namespace:    strings.TrimSpace(matches[1]),
				release:      strings.TrimSpace(matches[2]),
				resource:     strings.TrimSpace(matches[3]),
				resourceType: strings.TrimSpace(matches[4]),
				action:       strings.TrimSpace(matches[5]),
			})
		}

		if matches := helmSummaryRe.FindStringSubmatch(trimmed); len(matches) == 4 {
			fmt.Sscanf(matches[1], "%d", &summary.add)
			fmt.Sscanf(matches[2], "%d", &summary.change)
			fmt.Sscanf(matches[3], "%d", &summary.destroy)
		}
	}

	return changes, summary
}

func findHelmDiffForChange(change helmPlanChange, diffs []helmContentDiff) *helmContentDiff {
	for i := range diffs {
		diff := diffs[i]
		if !strings.EqualFold(strings.TrimSpace(string(diff.Kind)), strings.TrimSpace(change.resource)) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(string(diff.Name)), strings.TrimSpace(change.release)) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(string(diff.Namespace)), strings.TrimSpace(change.namespace)) {
			continue
		}
		if !helmResourceTypeMatches(string(diff.API), change.resourceType) {
			continue
		}

		return &diffs[i]
	}

	return nil
}

func helmResourceTypeMatches(api string, resourceType string) bool {
	left := strings.ToLower(strings.TrimSpace(api))
	right := strings.ToLower(strings.TrimSpace(resourceType))

	if left == "" || right == "" {
		return true
	}
	if left == right {
		return true
	}

	leftFamily := strings.Split(left, "/")[0]
	rightFamily := strings.Split(right, "/")[0]
	return leftFamily == rightFamily
}

func renderHelmDiffText(diff helmContentDiff) string {
	if isVersionTwo(diff.Version) {
		if renderedEntries := renderHelmDiffEntries(diff.Entries); renderedEntries != "" {
			return renderedEntries
		}
	}

	if renderedLines := renderHelmBeforeAfterDiff(string(diff.Before), string(diff.After)); renderedLines != "" {
		return renderedLines
	}

	return renderHelmDiffEntries(diff.Entries)
}

func isVersionTwo(version any) bool {
	switch v := version.(type) {
	case string:
		return strings.TrimSpace(v) == "2"
	case float64:
		return v == 2
	case int:
		return v == 2
	}

	return false
}

func renderHelmDiffEntries(entries []helmContentDiffEntry) string {
	if len(entries) == 0 {
		return ""
	}

	lines := []string{}
	for _, entry := range entries {
		entryBody := renderDiffEntryBody(entry)
		if entryBody == "" {
			continue
		}

		diffType := diffEntryType(entry.Delta, entry.Type)

		for _, line := range strings.Split(entryBody, "\n") {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" {
				continue
			}

			switch diffType {
			case 1:
				lines = append(lines, "- "+trimmedLine)
			case 2:
				lines = append(lines, "+ "+trimmedLine)
			case 0:
				lines = append(lines, "  "+trimmedLine)
			default:
				lines = append(lines, trimmedLine)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func renderDiffEntryBody(entry helmContentDiffEntry) string {
	path := strings.TrimSpace(string(entry.Path))
	payload := strings.TrimSpace(string(entry.Payload))

	if path == "" {
		return payload
	}
	if payload == "" {
		return path
	}

	return fmt.Sprintf("%s: %s", path, payload)
}

func diffEntryType(delta any, fallback any) int {
	if n, ok := coerceDiffTypeNumber(delta); ok {
		return n
	}

	if n, ok := coerceDiffTypeNumber(fallback); ok {
		return n
	}

	return 0
}

func coerceDiffTypeNumber(value any) (int, bool) {
	switch v := value.(type) {
	case nil:
		return 0, false
	case int:
		return v, true
	case float64:
		if v != math.Trunc(v) {
			return 0, false
		}
		return int(v), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func renderHelmBeforeAfterDiff(before string, after string) string {
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)

	if before == "" && after == "" {
		return ""
	}

	if before == "" {
		added := []string{}
		for _, line := range strings.Split(after, "\n") {
			added = append(added, "+ "+line)
		}
		return strings.Join(added, "\n")
	}

	if after == "" {
		removed := []string{}
		for _, line := range strings.Split(before, "\n") {
			removed = append(removed, "- "+line)
		}
		return strings.Join(removed, "\n")
	}

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	max := len(beforeLines)
	if len(afterLines) > max {
		max = len(afterLines)
	}

	lines := []string{}
	for i := 0; i < max; i++ {
		beforeLine := ""
		if i < len(beforeLines) {
			beforeLine = beforeLines[i]
		}
		afterLine := ""
		if i < len(afterLines) {
			afterLine = afterLines[i]
		}

		if beforeLine == afterLine {
			if afterLine == "" {
				continue
			}
			lines = append(lines, "  "+afterLine)
			continue
		}

		if beforeLine != "" {
			lines = append(lines, "- "+beforeLine)
		}
		if afterLine != "" {
			lines = append(lines, "+ "+afterLine)
		}
	}

	return strings.Join(lines, "\n")
}

func synthesizeHelmChanges(content []helmContentDiff) []helmPlanChange {
	changes := make([]helmPlanChange, 0, len(content))
	for _, diff := range content {
		release := strings.TrimSpace(string(diff.Name))
		if release == "" {
			release = strings.TrimSpace(string(diff.Resource))
		}

		resource := strings.TrimSpace(string(diff.Kind))
		if resource == "" {
			resource = strings.TrimSpace(string(diff.Resource))
		}

		changes = append(changes, helmPlanChange{
			namespace:    strings.TrimSpace(string(diff.Namespace)),
			release:      release,
			resource:     resource,
			resourceType: strings.TrimSpace(string(diff.API)),
			action:       inferHelmAction(diff),
		})
	}

	return changes
}

func inferHelmAction(diff helmContentDiff) string {
	if actionFromOp := strings.TrimSpace(string(diff.Op)); actionFromOp != "" {
		return normalizeHelmAction(actionFromOp)
	}

	hasBefore := strings.TrimSpace(string(diff.Before)) != ""
	hasAfter := strings.TrimSpace(string(diff.After)) != ""

	if len(diff.Entries) > 0 {
		hasRemoved := false
		hasAdded := false
		for _, entry := range diff.Entries {
			diffType := diffEntryType(entry.Delta, entry.Type)

			switch diffType {
			case 1:
				hasRemoved = true
			case 2:
				hasAdded = true
			}
		}

		if hasAdded && !hasRemoved {
			return "added"
		}
		if hasRemoved && !hasAdded {
			return "destroyed"
		}
		if hasAdded && hasRemoved {
			return "changed"
		}
	}

	if !hasBefore && hasAfter {
		return "added"
	}
	if hasBefore && !hasAfter {
		return "destroyed"
	}
	return "changed"
}

func incrementHelmSummaryByAction(summary *helmPlanSummary, action string) {
	switch normalizeHelmAction(action) {
	case "added":
		summary.add += 1
	case "destroyed":
		summary.destroy += 1
	default:
		summary.change += 1
	}
}

func normalizeHelmAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "add", "added", "create", "created":
		return "added"
	case "change", "changed", "update", "updated", "replace", "replaced", "apply", "patch", "patched", "configure", "configured":
		return "changed"
	case "destroy", "destroyed", "delete", "deleted", "remove", "removed":
		return "destroyed"
	default:
		return strings.ToLower(strings.TrimSpace(action))
	}
}

func isInteractivePlanDiffType(diffType componentDiffType) bool {
	return diffType == componentDiffTypeHelm || diffType == componentDiffTypeKubernetesManifest
}

func renderManifestContextSection(title string, body string) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.TextSubtle.Bold(true).Margin(1, 0, 0, 1).Render(title),
		body,
	)
}

func renderDiffHeaderRow(left string, right string, width int) string {
	contentWidth := width
	if contentWidth < 20 {
		contentWidth = 20
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	if leftWidth+rightWidth+1 > contentWidth {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", right)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		strings.Repeat(" ", contentWidth-leftWidth-rightWidth),
		right,
	)
}

func (m model) renderDiffText(text string) string {
	return m.renderDiffTextWithWidth(text, m.stepDetail.Width()-4)
}

func renderHelmDiffTextWithWidth(text string, width int) string {
	content := strings.Trim(text, "\n")
	if strings.TrimSpace(content) == "" {
		content = "No diff contents available"
	}

	content = collapseHelmDiffText(content)

	lines := strings.Split(content, "\n")
	styledLines := make([]string, 0, len(lines))
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			styledLines = append(styledLines, styles.TextSuccess.Render(line))
		case strings.HasPrefix(line, "-"):
			styledLines = append(styledLines, styles.TextError.Render(line))
		default:
			styledLines = append(styledLines, styles.TextDim.Render(line))
		}
	}

	if width < 20 {
		width = 20
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(width).
		Render(strings.Join(styledLines, "\n"))
}

func collapseHelmDiffText(text string) string {
	content := strings.Trim(text, "\n")
	if strings.TrimSpace(content) == "" {
		return text
	}

	rawLines := strings.Split(content, "\n")
	if len(rawLines) < helmDiffCollapseMinLines {
		return content
	}

	parsedLines := parseHelmRenderedDiffLines(rawLines)
	changedCount := 0
	for _, line := range parsedLines {
		if line.isChanged {
			changedCount += 1
		}
	}
	if changedCount == 0 {
		return content
	}

	keep := make([]bool, len(parsedLines))
	for i, line := range parsedLines {
		if !line.isChanged {
			continue
		}

		keep[i] = true

		for offset := 1; offset <= helmDiffContextLines; offset++ {
			if prev := i - offset; prev >= 0 && !parsedLines[prev].isChanged {
				keep[prev] = true
			}
			if next := i + offset; next < len(parsedLines) && !parsedLines[next].isChanged {
				keep[next] = true
			}
		}

		markHelmAncestorContext(keep, parsedLines, i)
	}

	collapsed, hiddenCount := buildCollapsedHelmDiffLines(parsedLines, keep)
	if hiddenCount < helmDiffCollapseMinHidden {
		return content
	}

	return strings.Join(collapsed, "\n")
}

func parseHelmRenderedDiffLines(lines []string) []helmRenderedDiffLine {
	parsed := make([]helmRenderedDiffLine, 0, len(lines))
	for _, raw := range lines {
		lineBody := raw
		isChanged := false
		switch {
		case strings.HasPrefix(raw, "+ "):
			lineBody = raw[2:]
			isChanged = true
		case strings.HasPrefix(raw, "- "):
			lineBody = raw[2:]
			isChanged = true
		case strings.HasPrefix(raw, "  "):
			lineBody = raw[2:]
		}

		trimmedBody := strings.TrimSpace(lineBody)
		parsed = append(parsed, helmRenderedDiffLine{
			raw:         raw,
			indent:      len(lineBody) - len(strings.TrimLeft(lineBody, " ")),
			isChanged:   isChanged,
			isContainer: strings.HasSuffix(trimmedBody, ":"),
		})
	}

	return parsed
}

func markHelmAncestorContext(keep []bool, lines []helmRenderedDiffLine, idx int) {
	if idx < 0 || idx >= len(lines) {
		return
	}

	targetIndent := lines[idx].indent
	if targetIndent <= 0 {
		return
	}

	for j := idx - 1; j >= 0; j-- {
		line := lines[j]
		if line.isChanged || !line.isContainer {
			continue
		}
		if line.indent >= targetIndent {
			continue
		}

		keep[j] = true
		targetIndent = line.indent
		if targetIndent == 0 {
			return
		}
	}
}

func buildCollapsedHelmDiffLines(lines []helmRenderedDiffLine, keep []bool) ([]string, int) {
	collapsed := make([]string, 0, len(lines))
	hiddenCount := 0

	for i := 0; i < len(lines); {
		if keep[i] {
			collapsed = append(collapsed, lines[i].raw)
			i += 1
			continue
		}

		rangeStart := i
		for i < len(lines) && !keep[i] {
			i += 1
		}
		rangeEnd := i - 1
		hiddenCount += rangeEnd - rangeStart + 1

		ellipsis := renderHelmEllipsisLine(lines, rangeStart, rangeEnd)
		if len(collapsed) == 0 || collapsed[len(collapsed)-1] != ellipsis {
			collapsed = append(collapsed, ellipsis)
		}
	}

	return collapsed, hiddenCount
}

func renderHelmEllipsisLine(lines []helmRenderedDiffLine, start int, end int) string {
	indent := 0
	hasIndent := false

	if start > 0 {
		indent = lines[start-1].indent
		hasIndent = true
	}
	if end+1 < len(lines) {
		nextIndent := lines[end+1].indent
		if !hasIndent || nextIndent < indent {
			indent = nextIndent
			hasIndent = true
		}
	}
	if !hasIndent {
		indent = 0
	}

	return "  " + strings.Repeat(" ", indent) + "..."
}

func (m model) renderDiffTextWithWidth(text string, width int) string {
	content := strings.TrimSpace(text)
	if content == "" {
		content = "No diff contents available"
	}

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

func extractDryRunOutput(value any) string {
	if value == nil {
		return ""
	}

	if text := findDryRunOutput(value); text != "" {
		return text
	}

	return ""
}

func findDryRunOutput(value any) string {
	switch v := value.(type) {
	case map[string]any:
		for k, raw := range v {
			if strings.EqualFold(k, "dry_run_output") {
				if text := coerceDisplayDiffText(raw); text != "" {
					return text
				}
			}
		}

		if rawPlan, ok := v["plan"]; ok {
			if planText := coerceDisplayDiffText(rawPlan); looksLikeJSONPayload(planText) {
				var nested any
				if err := json.Unmarshal([]byte(planText), &nested); err == nil {
					if text := findDryRunOutput(nested); text != "" {
						return text
					}
				}
			}
		}

		for _, raw := range v {
			if text := findDryRunOutput(raw); text != "" {
				return text
			}
		}
	case []any:
		for _, raw := range v {
			if text := findDryRunOutput(raw); text != "" {
				return text
			}
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if !looksLikeJSONPayload(trimmed) {
			return ""
		}

		var nested any
		if err := json.Unmarshal([]byte(trimmed), &nested); err != nil {
			return ""
		}

		return findDryRunOutput(nested)
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
