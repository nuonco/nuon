package workflow

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

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
