package activities

import (
	"testing"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/stretchr/testify/require"
)

func TestDiffNodeToSection(t *testing.T) {
	t.Run("nil node returns nil", func(t *testing.T) {
		result := diffNodeToSection(nil)
		require.Nil(t, result)
	})

	t.Run("unknown key returns nil", func(t *testing.T) {
		node := &diff.Diff{
			Key: "unknown_section",
			Children: []*diff.Diff{
				{Key: "item", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "added"}},
			},
		}
		result := diffNodeToSection(node)
		require.Nil(t, result)
	})

	t.Run("components section with adds and removes", func(t *testing.T) {
		node := &diff.Diff{
			Key: "components",
			Children: []*diff.Diff{
				{Key: "api", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "new component"}},
				{Key: "worker", Diff: &diff.DiffKey{Op: diff.OpRemove, Diff: "removed component"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, "Components", section.Name)
		require.Equal(t, 1, section.Additions)
		require.Equal(t, 1, section.Removals)
		require.Equal(t, 0, section.Changed)
		require.Len(t, section.Entries, 2)
	})

	t.Run("actions section with changes", func(t *testing.T) {
		node := &diff.Diff{
			Key: "actions",
			Children: []*diff.Diff{
				{Key: "deploy", Diff: &diff.DiffKey{Op: diff.OpChange, Diff: "modified"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, "Actions", section.Name)
		require.Equal(t, 0, section.Additions)
		require.Equal(t, 0, section.Removals)
		require.Equal(t, 1, section.Changed)
	})

	t.Run("no-op entries filtered out", func(t *testing.T) {
		node := &diff.Diff{
			Key: "components",
			Children: []*diff.Diff{
				{Key: "api", Diff: &diff.DiffKey{Op: diff.OpNoop, Diff: "unchanged"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Empty(t, section.Entries)
	})

	t.Run("empty children produces no entries", func(t *testing.T) {
		node := &diff.Diff{
			Key:      "components",
			Children: []*diff.Diff{},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Empty(t, section.Entries)
	})

	t.Run("ungrouped section counts as single entity", func(t *testing.T) {
		node := &diff.Diff{
			Key: "sandbox",
			Children: []*diff.Diff{
				{Key: "terraform_version", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> '1.5.0'"}},
				{Key: "drift_schedule", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> '0 * * * *'"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, "Sandbox", section.Name)
		require.Equal(t, 1, section.Additions)
		require.Equal(t, 0, section.Removals)
		require.Equal(t, 0, section.Changed)
		require.Len(t, section.Entries, 1)
	})

	t.Run("ungrouped section with adds and changes but no removes counts as add", func(t *testing.T) {
		node := &diff.Diff{
			Key: "runner",
			Children: []*diff.Diff{
				{Key: "runner_type", Diff: &diff.DiffKey{Op: diff.OpChange, Diff: "'false' -> 'true'"}},
				{Key: "helm_driver", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> 'secrets'"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, "Runner", section.Name)
		require.Equal(t, 1, section.Additions)
		require.Equal(t, 0, section.Removals)
		require.Equal(t, 0, section.Changed)
	})

	t.Run("ungrouped section with adds and removes counts as change", func(t *testing.T) {
		node := &diff.Diff{
			Key: "runner",
			Children: []*diff.Diff{
				{Key: "runner_type", Diff: &diff.DiffKey{Op: diff.OpRemove, Diff: "'eks' -> ''"}},
				{Key: "helm_driver", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> 'secrets'"}},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, "Runner", section.Name)
		require.Equal(t, 0, section.Additions)
		require.Equal(t, 0, section.Removals)
		require.Equal(t, 1, section.Changed)
	})

	t.Run("grouped entity with nested children counts correctly", func(t *testing.T) {
		node := &diff.Diff{
			Key: "components",
			Children: []*diff.Diff{
				{
					Key: "component.api",
					Children: []*diff.Diff{
						{Key: "type", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> 'helm_chart'"}},
						{Key: "var_name", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "'' -> 'api'"}},
					},
				},
			},
		}

		section := diffNodeToSection(node)
		require.NotNil(t, section)
		require.Equal(t, 1, section.Additions)
		require.Equal(t, 0, section.Changed)
		require.Len(t, section.Entries, 1)
	})
}

func TestEntityAggregateOp(t *testing.T) {
	t.Run("nil node returns empty", func(t *testing.T) {
		require.Equal(t, diff.Op(""), entityAggregateOp(nil))
	})

	t.Run("all adds returns add", func(t *testing.T) {
		node := &diff.Diff{
			Children: []*diff.Diff{
				{Key: "a", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "added"}},
				{Key: "b", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "added"}},
			},
		}
		require.Equal(t, diff.OpAdd, entityAggregateOp(node))
	})

	t.Run("all removes returns remove", func(t *testing.T) {
		node := &diff.Diff{
			Children: []*diff.Diff{
				{Key: "a", Diff: &diff.DiffKey{Op: diff.OpRemove, Diff: "removed"}},
			},
		}
		require.Equal(t, diff.OpRemove, entityAggregateOp(node))
	})

	t.Run("adds with changes but no removes returns add", func(t *testing.T) {
		node := &diff.Diff{
			Children: []*diff.Diff{
				{Key: "a", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "added"}},
				{Key: "b", Diff: &diff.DiffKey{Op: diff.OpChange, Diff: "'false' -> 'true'"}},
			},
		}
		require.Equal(t, diff.OpAdd, entityAggregateOp(node))
	})

	t.Run("adds and removes returns change", func(t *testing.T) {
		node := &diff.Diff{
			Children: []*diff.Diff{
				{Key: "a", Diff: &diff.DiffKey{Op: diff.OpAdd, Diff: "added"}},
				{Key: "b", Diff: &diff.DiffKey{Op: diff.OpRemove, Diff: "removed"}},
			},
		}
		require.Equal(t, diff.OpChange, entityAggregateOp(node))
	})

	t.Run("only noops returns empty", func(t *testing.T) {
		node := &diff.Diff{
			Children: []*diff.Diff{
				{Key: "a", Diff: &diff.DiffKey{Op: diff.OpNoop, Diff: "unchanged"}},
			},
		}
		require.Equal(t, diff.Op(""), entityAggregateOp(node))
	})
}

func TestSectionDisplayNameAndGrouped(t *testing.T) {
	tests := []struct {
		key             string
		expectedName    string
		expectedGrouped bool
	}{
		{"components", "Components", true},
		{"actions", "Actions", true},
		{"inputs", "Install inputs", true},
		{"secrets", "Secrets", true},
		{"policies", "Policies", true},
		{"sandbox", "Sandbox", false},
		{"runner", "Runner", false},
		{"permissions", "Permissions", false},
		{"stack", "Stack", false},
		{"break_glass", "Break glass", false},
		{"operation_roles", "Operation roles", false},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			name, grouped := sectionDisplayNameAndGrouped(tt.key)
			require.Equal(t, tt.expectedName, name)
			require.Equal(t, tt.expectedGrouped, grouped)
		})
	}
}
