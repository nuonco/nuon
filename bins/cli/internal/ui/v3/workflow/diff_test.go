package workflow

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var ansiSeqRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func testHelmPayload() map[string]any {
	return map[string]any{
		"plan": "\u001b[0;32mwhoami, whoami, Service (v1) to be added.\u001b[0m\n" +
			"\u001b[0;33mwhoami, whoami, Deployment (apps) to be changed.\u001b[0m\n" +
			"Plan: 1 to add, 1 to change, 0 to destroy.\n",
		"op": "install",
		"helm_content_diff": []any{
			map[string]any{
				"api":       "v1",
				"name":      "whoami",
				"namespace": "whoami",
				"kind":      "Service",
				"before":    "",
				"after":     "apiVersion: v1\nkind: Service",
			},
			map[string]any{
				"api":       "apps/v1",
				"name":      "whoami",
				"namespace": "whoami",
				"kind":      "Deployment",
				"before":    "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: whoami",
				"after":     "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: whoami-v2",
			},
		},
	}
}

func testKubernetesManifestPayload() map[string]any {
	nestedPlan := map[string]any{
		"op": "apply",
		"k8s_content_diff": []any{
			map[string]any{
				"api":       "v1",
				"name":      "install-secrets",
				"namespace": "nuon-system",
				"kind":      "Secret",
				"resource":  "secrets",
				"op":        "patch",
				"entries": []any{
					map[string]any{
						"path":    "metadata.labels.app",
						"type":    0,
						"payload": "nuon",
					},
					map[string]any{
						"path":    "data.INSTALL_ID",
						"type":    2,
						"payload": "inst_123",
					},
				},
			},
		},
		"dry_run_output": "apiVersion: v1\nkind: Secret\nmetadata:\n  name: install-secrets",
	}

	nestedPlanBytes, _ := json.Marshal(nestedPlan)

	return map[string]any{
		"op":   "apply",
		"plan": string(nestedPlanBytes),
	}
}

func TestExtractDisplayDiffText(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name: "extracts nested apply plan display text",
			input: map[string]any{
				"deploy_plan": map[string]any{
					"apply_plan_display": "- old\n+ new",
				},
			},
			want: "- old\n+ new",
		},
		{
			name: "extracts plan display bytes encoded as array",
			input: map[string]any{
				"sandbox_mode": map[string]any{
					"plan_display_contents": []any{45.0, 32.0, 111.0, 108.0, 100.0},
				},
			},
			want: "- old",
		},
		{
			name:  "falls back to plain string payload",
			input: "- before\n+ after",
			want:  "- before\n+ after",
		},
		{
			name:  "returns empty when no display fields exist",
			input: map[string]any{"foo": "bar"},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDisplayDiffText(tt.input)
			if got != tt.want {
				t.Fatalf("extractDisplayDiffText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectTerraformDiffGroupsIncludesResourceDrift(t *testing.T) {
	plan := tfjson.Plan{
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "module.foo.aws_db_instance.this[0]",
				Change:  &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionNoop}},
			},
		},
		ResourceDrift: []*tfjson.ResourceChange{
			{
				Address: "module.foo.aws_db_instance.this[0]",
				Change:  &tfjson.Change{Actions: tfjson.Actions{tfjson.ActionUpdate}},
			},
		},
	}

	groups := collectTerraformDiffGroups(plan)
	if len(groups.updates) != 1 {
		t.Fatalf("expected 1 update from resource_drift, got %d", len(groups.updates))
	}

	if len(groups.creations) != 0 || len(groups.deletions) != 0 {
		t.Fatalf("expected no create/delete changes, got create=%d delete=%d", len(groups.creations), len(groups.deletions))
	}
}

func TestSelectedStepDiffType(t *testing.T) {
	tests := []struct {
		name string
		m    model
		want componentDiffType
	}{
		{
			name: "uses terraform approval type",
			m: model{selectedStep: &models.AppWorkflowStep{Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeTerraformPlan,
			}}},
			want: componentDiffTypeTerraform,
		},
		{
			name: "uses helm approval type",
			m: model{selectedStep: &models.AppWorkflowStep{Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeHelmApproval,
			}}},
			want: componentDiffTypeHelm,
		},
		{
			name: "uses kubernetes manifest approval type",
			m: model{selectedStep: &models.AppWorkflowStep{Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeKubernetesManifestApproval,
			}}},
			want: componentDiffTypeKubernetesManifest,
		},
		{
			name: "falls back to terraform plan key detection",
			m: model{approvalContents: approvalContents{contents: map[string]interface{}{
				"terraform_version": "1.6.5",
			}}},
			want: componentDiffTypeTerraform,
		},
		{
			name: "falls back to helm key detection",
			m: model{approvalContents: approvalContents{contents: map[string]interface{}{
				"helm_content_diff": []any{},
			}}},
			want: componentDiffTypeHelm,
		},
		{
			name: "falls back to kubernetes key detection",
			m: model{approvalContents: approvalContents{contents: map[string]interface{}{
				"k8s_content_diff": []any{},
			}}},
			want: componentDiffTypeKubernetesManifest,
		},
		{
			name: "defaults to unknown",
			m:    model{},
			want: componentDiffTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.selectedStepDiffType()
			if got != tt.want {
				t.Fatalf("selectedStepDiffType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseHelmPlanText(t *testing.T) {
	planText := "\u001b[0;32mwhoami, whoami, Deployment (apps) to be added.\u001b[0m\nPlan: 1 to add, 0 to change, 0 to destroy.\n"

	changes, summary := parseHelmPlanText(planText)
	if len(changes) != 1 {
		t.Fatalf("expected 1 parsed change, got %d", len(changes))
	}

	change := changes[0]
	if change.namespace != "whoami" || change.release != "whoami" || change.resource != "Deployment" || change.resourceType != "apps" || change.action != "added" {
		t.Fatalf("unexpected parsed change: %+v", change)
	}

	if summary.add != 1 || summary.change != 0 || summary.destroy != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestParseHelmPlanSupportsNestedJSONPlanPayload(t *testing.T) {
	parsed, err := parseHelmPlan(testKubernetesManifestPayload())
	if err != nil {
		t.Fatalf("expected parseHelmPlan to parse nested plan payload, got error: %v", err)
	}

	if parsed.planSource != "kubernetes" {
		t.Fatalf("expected kubernetes plan source, got %q", parsed.planSource)
	}

	if len(parsed.content) != 1 {
		t.Fatalf("expected 1 kubernetes diff entry, got %d", len(parsed.content))
	}

	if len(parsed.changes) != 1 {
		t.Fatalf("expected 1 synthesized change, got %d", len(parsed.changes))
	}

	if parsed.changes[0].action != "changed" {
		t.Fatalf("expected action inferred from op=patch to be changed, got %q", parsed.changes[0].action)
	}

	if !strings.Contains(parsed.dryRun, "kind: Secret") {
		t.Fatalf("expected dry run output from nested payload, got %q", parsed.dryRun)
	}
}

func TestFindHelmDiffForChange(t *testing.T) {
	change := helmPlanChange{
		namespace:    "default",
		release:      "whoami",
		resource:     "Deployment",
		resourceType: "apps",
		action:       "added",
	}

	diffs := []helmContentDiff{
		{
			API:       "apps/v1",
			Name:      "whoami",
			Namespace: "default",
			Kind:      "Deployment",
			After:     "apiVersion: apps/v1",
		},
	}

	matched := findHelmDiffForChange(change, diffs)
	if matched == nil {
		t.Fatalf("expected matching helm diff")
	}

	if !strings.Contains(string(matched.After), "apps/v1") {
		t.Fatalf("expected matched diff contents, got %+v", matched)
	}
}

func TestRenderHelmDiffEntries(t *testing.T) {
	entries := []helmContentDiffEntry{
		{Type: 1, Payload: "old: value"},
		{Type: 2, Payload: "new: value"},
		{Type: 2, Path: "data.INSTALL_ID", Payload: "inst_123"},
		{Payload: "unchanged: value"},
	}

	rendered := renderHelmDiffEntries(entries)
	if !strings.Contains(rendered, "- old: value") {
		t.Fatalf("expected removed line in rendered diff, got %q", rendered)
	}
	if !strings.Contains(rendered, "+ new: value") {
		t.Fatalf("expected added line in rendered diff, got %q", rendered)
	}
	if !strings.Contains(rendered, "+ data.INSTALL_ID: inst_123") {
		t.Fatalf("expected path-based added line in rendered diff, got %q", rendered)
	}
	if !strings.Contains(rendered, "  unchanged: value") {
		t.Fatalf("expected unchanged line in rendered diff, got %q", rendered)
	}
}

func TestStepDetailViewHelmDiff(t *testing.T) {
	raw := testHelmPayload()

	m := model{
		approvalContents: approvalContents{raw: raw, contents: raw},
		selectedStep: &models.AppWorkflowStep{
			ID:             "step-1",
			StepTargetType: "install_deploys",
			Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeHelmApproval,
			},
		},
		stepDetail: viewport.New(viewport.WithWidth(120), viewport.WithHeight(30)),
	}
	view := ansiSeqRe.ReplaceAllString(m.stepDetailViewHelmDiff(), "")

	if !strings.Contains(view, "Helm changes overview") {
		t.Fatalf("expected helm overview heading in view, got %q", view)
	}
	if !strings.Contains(view, "Operation: install") {
		t.Fatalf("expected operation in view, got %q", view)
	}
	if !strings.Contains(view, "Use [j/k] to select, [enter] to expand, and [↑/↓] to scroll.") {
		t.Fatalf("expected interactive hint in view, got %q", view)
	}
	if !strings.Contains(view, "▸ whoami") {
		t.Fatalf("expected collapsed indicator in view, got %q", view)
	}
	if !strings.Contains(view, "whoami") || !strings.Contains(view, "Service") {
		t.Fatalf("expected parsed change metadata in view, got %q", view)
	}
	if strings.Contains(view, "+ apiVersion: v1") {
		t.Fatalf("expected collapsed diff body by default, got %q", view)
	}

	if !m.helmDiffExplorer.toggleSelectedExpanded() {
		t.Fatalf("expected toggleSelectedExpanded to expand selected row")
	}

	expandedView := ansiSeqRe.ReplaceAllString(m.stepDetailViewHelmDiff(), "")
	if !strings.Contains(expandedView, "▾ whoami") {
		t.Fatalf("expected expanded indicator in view, got %q", expandedView)
	}
	if !strings.Contains(expandedView, "+ apiVersion: v1") {
		t.Fatalf("expected rendered diff text in expanded view, got %q", expandedView)
	}
}

func TestHelmDiffExplorerSelectionAndExpansionState(t *testing.T) {
	raw := testHelmPayload()

	explorer := newHelmDiffExplorerModel(100)
	explorer.Bind("step-1", raw, raw)

	if !explorer.hasInteractiveRows() {
		t.Fatalf("expected explorer to have interactive rows")
	}
	if explorer.selectedIndex != 0 {
		t.Fatalf("expected initial selected index 0, got %d", explorer.selectedIndex)
	}

	if !explorer.moveSelection(1) {
		t.Fatalf("expected moveSelection to move to the next row")
	}
	if explorer.selectedIndex != 1 {
		t.Fatalf("expected selected index 1, got %d", explorer.selectedIndex)
	}

	if !explorer.toggleSelectedExpanded() {
		t.Fatalf("expected toggleSelectedExpanded to expand current row")
	}
	if !explorer.expanded[1] {
		t.Fatalf("expected second row to be expanded")
	}
	if explorer.expanded[0] {
		t.Fatalf("expected accordion behavior to collapse non-selected rows")
	}

	explorer.Bind("step-1", raw, raw)
	if explorer.selectedIndex != 1 {
		t.Fatalf("expected selection to be preserved when rebinding same step, got %d", explorer.selectedIndex)
	}
	if !explorer.expanded[1] {
		t.Fatalf("expected expanded row to be preserved when rebinding same step")
	}

	explorer.Bind("step-2", raw, raw)
	if explorer.selectedIndex != 0 {
		t.Fatalf("expected selection reset when switching steps, got %d", explorer.selectedIndex)
	}
	if len(explorer.expanded) != 0 {
		t.Fatalf("expected expanded rows reset when switching steps")
	}
}

func TestHandleDetailContentKeyRoutesToHelmDiffExplorer(t *testing.T) {
	raw := testHelmPayload()

	m := model{
		focus: "detail",
		selectedStep: &models.AppWorkflowStep{
			ID:             "step-1",
			StepTargetType: "install_deploys",
			Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeHelmApproval,
			},
		},
		approvalContents: approvalContents{raw: raw, contents: raw},
		stepDetail:       viewport.New(viewport.WithWidth(120), viewport.WithHeight(30)),
	}

	m.syncHelmDiffExplorer()

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: tea.KeyDown}); handled {
		t.Fatalf("expected down key to be left for viewport scrolling")
	}
	if m.helmDiffExplorer.selectedIndex != 0 {
		t.Fatalf("expected selected diff index to remain unchanged on down arrow, got %d", m.helmDiffExplorer.selectedIndex)
	}

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: 'j', Text: "j"}); !handled {
		t.Fatalf("expected j key to be handled by detail content model")
	}
	if m.helmDiffExplorer.selectedIndex != 1 {
		t.Fatalf("expected selected diff index to move to 1, got %d", m.helmDiffExplorer.selectedIndex)
	}

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: tea.KeyEnter}); !handled {
		t.Fatalf("expected enter key to be handled by detail content model")
	}
	if !m.helmDiffExplorer.expanded[1] {
		t.Fatalf("expected selected diff card to be expanded")
	}
	if m.helmDiffExplorer.expanded[0] {
		t.Fatalf("expected only one expanded diff card at a time")
	}
}

func TestStepDetailViewKubernetesManifestDiff(t *testing.T) {
	raw := testKubernetesManifestPayload()

	m := model{
		approvalContents: approvalContents{raw: raw, contents: raw},
		selectedStep: &models.AppWorkflowStep{
			ID:             "step-k8s-1",
			StepTargetType: "install_deploys",
			Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeKubernetesManifestApproval,
			},
		},
		stepDetail: viewport.New(viewport.WithWidth(120), viewport.WithHeight(30)),
	}

	view := ansiSeqRe.ReplaceAllString(m.stepDetailViewKubernetesManifestDiff(), "")

	if !strings.Contains(view, "Kubernetes changes overview") {
		t.Fatalf("expected kubernetes overview heading in view, got %q", view)
	}
	if !strings.Contains(view, "Operation: apply") {
		t.Fatalf("expected operation in view, got %q", view)
	}
	if !strings.Contains(view, "▸ install-secrets") {
		t.Fatalf("expected collapsed indicator in view, got %q", view)
	}
	if !strings.Contains(view, "Secret") || !strings.Contains(view, "nuon-system") {
		t.Fatalf("expected parsed kubernetes metadata in view, got %q", view)
	}
	if strings.Contains(view, "+ data.INSTALL_ID: inst_123") {
		t.Fatalf("expected collapsed diff body by default, got %q", view)
	}

	if !m.helmDiffExplorer.toggleSelectedExpanded() {
		t.Fatalf("expected toggleSelectedExpanded to expand selected kubernetes row")
	}

	expandedView := ansiSeqRe.ReplaceAllString(m.stepDetailViewKubernetesManifestDiff(), "")
	if !strings.Contains(expandedView, "▾ install-secrets") {
		t.Fatalf("expected expanded indicator in view, got %q", expandedView)
	}
	if !strings.Contains(expandedView, "+ data.INSTALL_ID: inst_123") {
		t.Fatalf("expected rendered path-based diff in expanded view, got %q", expandedView)
	}
}

func TestKubernetesManifestDiffFallsBackToDryRunOutput(t *testing.T) {
	raw := map[string]any{
		"op":             "apply",
		"dry_run_output": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: from-dry-run",
	}

	m := model{
		approvalContents: approvalContents{raw: raw, contents: raw},
		selectedStep: &models.AppWorkflowStep{
			ID:             "step-k8s-dry-run",
			StepTargetType: "install_deploys",
			Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeKubernetesManifestApproval,
			},
		},
		stepDetail: viewport.New(viewport.WithWidth(120), viewport.WithHeight(30)),
	}

	view := ansiSeqRe.ReplaceAllString(m.stepDetailViewKubernetesManifestDiff(), "")

	if !strings.Contains(view, "Dry-run manifest output") {
		t.Fatalf("expected dry-run section heading, got %q", view)
	}
	if !strings.Contains(view, "kind: ConfigMap") {
		t.Fatalf("expected dry-run manifest YAML in fallback view, got %q", view)
	}
}

func TestHandleDetailContentKeyRoutesToKubernetesDiffExplorer(t *testing.T) {
	raw := testKubernetesManifestPayload()

	m := model{
		focus: "detail",
		selectedStep: &models.AppWorkflowStep{
			ID:             "step-k8s-keys",
			StepTargetType: "install_deploys",
			Approval: &models.AppWorkflowStepApproval{
				Type: models.AppWorkflowStepApprovalTypeKubernetesManifestApproval,
			},
		},
		approvalContents: approvalContents{raw: raw, contents: raw},
		stepDetail:       viewport.New(viewport.WithWidth(120), viewport.WithHeight(30)),
	}

	m.syncHelmDiffExplorer()

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: tea.KeyDown}); handled {
		t.Fatalf("expected down key to be left for viewport scrolling")
	}

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: 'j', Text: "j"}); !handled {
		t.Fatalf("expected j key to be consumed by kubernetes diff explorer")
	}
	if m.helmDiffExplorer.selectedIndex != 0 {
		t.Fatalf("expected selected index to remain at 0 for a single row, got %d", m.helmDiffExplorer.selectedIndex)
	}

	if handled := m.handleDetailContentKey(tea.KeyPressMsg{Code: tea.KeyEnter}); !handled {
		t.Fatalf("expected enter key to be handled by kubernetes diff explorer")
	}

	if !m.helmDiffExplorer.expanded[0] {
		t.Fatalf("expected selected kubernetes diff card to be expanded")
	}
}

func TestRenderHelmDiffTextWithWidth(t *testing.T) {
	rendered := ansiSeqRe.ReplaceAllString(renderHelmDiffTextWithWidth("  unchanged\n- old\n+ new", 80), "")
	if !strings.Contains(rendered, "  unchanged") || !strings.Contains(rendered, "- old") || !strings.Contains(rendered, "+ new") {
		t.Fatalf("expected rendered diff text to include all lines, got %q", rendered)
	}
}

func TestCollapseHelmDiffTextPreservesContextAndNesting(t *testing.T) {
	input := strings.Join([]string{
		"  toplevelattr:",
		"    subsection:",
		"      first_line: lol",
		"      second_line: lol",
		"      third_line: lol",
		"      change_minus_1: lol",
		"-     change_line: lol",
		"+     change_line: lmao",
		"      change_plus_1: lol",
		"      fourth_line: lol",
		"      fifth_line: lol",
		"  sibling:",
		"    value: keep",
	}, "\n")

	collapsed := collapseHelmDiffText(input)

	if !strings.Contains(collapsed, "  toplevelattr:") || !strings.Contains(collapsed, "    subsection:") {
		t.Fatalf("expected collapsed output to retain nesting keys, got %q", collapsed)
	}
	if !strings.Contains(collapsed, "      change_minus_1: lol") || !strings.Contains(collapsed, "      change_plus_1: lol") {
		t.Fatalf("expected collapsed output to retain one line of change context, got %q", collapsed)
	}
	if !strings.Contains(collapsed, "-     change_line: lol") || !strings.Contains(collapsed, "+     change_line: lmao") {
		t.Fatalf("expected collapsed output to retain changed lines, got %q", collapsed)
	}
	if !strings.Contains(collapsed, "...") {
		t.Fatalf("expected collapsed output to include ellipsis for hidden ranges, got %q", collapsed)
	}
	if strings.Contains(collapsed, "      first_line: lol") || strings.Contains(collapsed, "      fourth_line: lol") {
		t.Fatalf("expected collapsed output to omit distant unchanged lines, got %q", collapsed)
	}
}

func TestCollapseHelmDiffTextSkipsShortDiffs(t *testing.T) {
	input := strings.Join([]string{
		"  metadata:",
		"    name: whoami",
		"    labels:",
		"      app: whoami",
		"-     version: old",
		"+     version: new",
		"      stable: true",
	}, "\n")

	collapsed := collapseHelmDiffText(input)
	if collapsed != input {
		t.Fatalf("expected short diff to remain unchanged, got %q", collapsed)
	}
}
