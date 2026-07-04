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
}

func TestCollectEntries(t *testing.T) {
	t.Run("nested children attributed to top-level item", func(t *testing.T) {
		node := &diff.Diff{
			Key: "ctl-api",
			Children: []*diff.Diff{
				{
					Key: "env",
					Children: []*diff.Diff{
						{Key: "DATABASE_URL", Diff: &diff.DiffKey{Op: diff.OpChange, Diff: "old -> new"}},
					},
				},
			},
		}

		section := &ConfigDiffSection{Name: "Components"}
		collectEntries(node, "", section)

		require.Len(t, section.Entries, 1)
		require.Equal(t, "ctl-api", section.Entries[0].Name)
		require.Contains(t, section.Entries[0].Description, "DATABASE_URL")
		require.Equal(t, 1, section.Changed)
	})

	t.Run("nil node is safe", func(t *testing.T) {
		section := &ConfigDiffSection{Name: "test"}
		collectEntries(nil, "", section)
		require.Empty(t, section.Entries)
	})
}

func TestSectionDisplayName(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"components", "Components"},
		{"actions", "Actions"},
		{"inputs", "Install inputs"},
		{"secrets", "Secrets"},
		{"sandbox", "Sandbox"},
		{"runner", "Runner"},
		{"permissions", "Permissions"},
		{"stack", "Stack"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			require.Equal(t, tt.expected, sectionDisplayName(tt.key))
		})
	}
}
