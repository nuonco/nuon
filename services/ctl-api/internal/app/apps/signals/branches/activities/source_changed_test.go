package activities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRepoPath(t *testing.T) {
	require.Equal(t, "components/api", normalizeRepoPath("./components/api/"))
	require.Equal(t, "", normalizeRepoPath("."))
	require.Equal(t, "", normalizeRepoPath("./"))
	require.Equal(t, "foo", normalizeRepoPath("foo"))
}

func TestPathMatchesDirectory(t *testing.T) {
	require.True(t, pathMatchesDirectory("components/api/main.go", "components/api"))
	require.True(t, pathMatchesDirectory("components/api", "components/api"))
	require.False(t, pathMatchesDirectory("components/api-extra/main.go", "components/api"))
	require.False(t, pathMatchesDirectory("other/file.go", "components/api"))
	require.True(t, pathMatchesDirectory("anything/file.go", "."))
	require.True(t, pathMatchesDirectory("anything/file.go", ""))
	require.False(t, pathMatchesDirectory("", "components/api"))
}

func TestEnrichConfigDiffWithSourceChanged(t *testing.T) {
	full := &ComputeAppConfigDiffOutput{
		ConfigFile: "nuon.toml",
		Additions:  1,
		Changed:    1,
		Sections: []ConfigDiffSection{
			{
				Name:    "Components",
				Changed: 1,
				Entries: []ConfigDiffEntry{
					{Op: "change", Name: "api"},
					{Op: "add", Name: "worker"},
				},
			},
			{
				Name:      "Sandbox",
				Additions: 1,
				Entries: []ConfigDiffEntry{
					{Op: "add", Name: "sandbox"},
				},
			},
		},
	}

	dirs := map[string]string{
		"api":    "components/api",
		"worker": "components/worker",
	}
	changed := []string{"components/api/main.go", "docs/readme.md"}

	out := enrichConfigDiffWithSourceChanged(full, dirs, changed)
	require.Len(t, out.Sections, 2)

	comp := out.Sections[0]
	require.Equal(t, "Components", comp.Name)
	require.True(t, comp.Entries[0].SourceChanged)
	require.False(t, comp.Entries[1].SourceChanged)

	sandbox := out.Sections[1]
	require.Equal(t, "Sandbox", sandbox.Name)
	require.False(t, sandbox.Entries[0].SourceChanged)
}
