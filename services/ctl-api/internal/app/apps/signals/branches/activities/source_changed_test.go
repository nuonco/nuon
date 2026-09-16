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

func TestNormalizeRepoURL(t *testing.T) {
	require.True(t, repoURLsEqual("acme/app", "https://github.com/acme/app.git"))
	require.True(t, repoURLsEqual("git@github.com:acme/app.git", "acme/app"))
	require.False(t, repoURLsEqual("acme/app", "acme/other"))
	require.False(t, repoURLsEqual("", "acme/app"))
}

func TestPathMatchesDirectory(t *testing.T) {
	require.True(t, pathMatchesDirectory("components/api/main.go", "components/api"))
	require.True(t, pathMatchesDirectory("components/api", "components/api"))
	require.False(t, pathMatchesDirectory("components/api-extra/main.go", "components/api"))
	require.False(t, pathMatchesDirectory("other/file.go", "components/api"))
	require.True(t, pathMatchesDirectory("anything/file.go", "."))
	require.True(t, pathMatchesDirectory("anything/file.go", ""))
	require.False(t, pathMatchesDirectory("", "components/api"))
	require.False(t, pathMatchesDirectory("inputs/dns/domain.toml", "src/components/alb"))
}

func TestComponentSourceChanged(t *testing.T) {
	branch := "acme/app"
	changed := []string{"components/api/main.go", "docs/readme.md"}

	require.True(t, componentSourceChanged(componentSource{
		Name: "api", Repo: "https://github.com/acme/app.git", Directory: "components/api",
	}, branch, changed))

	require.False(t, componentSourceChanged(componentSource{
		Name: "worker", Repo: "acme/app", Directory: "components/worker",
	}, branch, changed))

	require.False(t, componentSourceChanged(componentSource{
		Name: "api", Repo: "acme/other", Directory: "components/api",
	}, branch, changed))

	require.True(t, componentSourceChanged(componentSource{
		Name: "root", Repo: "acme/app", Directory: ".",
	}, branch, []string{"inputs/dns/domain.toml"}))

	require.False(t, componentSourceChanged(componentSource{
		Name: "root", Repo: "acme/charts", Directory: ".",
	}, branch, []string{"inputs/dns/domain.toml"}))
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

	sources := []componentSource{
		{Name: "api", Repo: "acme/app", Directory: "components/api"},
		{Name: "worker", Repo: "acme/app", Directory: "components/worker"},
		{Name: "charts", Repo: "acme/charts", Directory: "."},
	}
	changed := []string{"components/api/main.go", "docs/readme.md"}

	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", changed)
	require.Len(t, out.Sections, 2)

	comp := out.Sections[0]
	require.Equal(t, "Components", comp.Name)
	require.True(t, comp.Entries[0].SourceChanged)
	require.False(t, comp.Entries[1].SourceChanged)

	sandbox := out.Sections[1]
	require.Equal(t, "Sandbox", sandbox.Name)
	require.False(t, sandbox.Entries[0].SourceChanged)

	require.True(t, out.ComponentSourceChanged["api"])
	require.False(t, out.ComponentSourceChanged["worker"])
	require.False(t, out.ComponentSourceChanged["charts"])
}

func TestEnrichConfigDiffSourceOnlyComponent(t *testing.T) {
	full := &ComputeAppConfigDiffOutput{
		Sections: []ConfigDiffSection{
			{Name: "Inputs", Entries: []ConfigDiffEntry{{Op: "change", Name: "dns"}}},
		},
	}
	sources := []componentSource{
		{Name: "api", Repo: "acme/app", Directory: "components/api"},
	}
	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", []string{"components/api/main.go"})
	require.True(t, out.ComponentSourceChanged["api"])
	require.False(t, out.Sections[0].Entries[0].SourceChanged)
}

func TestEnrichConfigDiffWithSourceChangedMissingDirectoryIsFalse(t *testing.T) {
	full := &ComputeAppConfigDiffOutput{
		Sections: []ConfigDiffSection{
			{
				Name: "Components",
				Entries: []ConfigDiffEntry{
					{Op: "change", Name: "manifest"},
				},
			},
		},
	}

	out := enrichConfigDiffWithSourceChanged(full, nil, "acme/app", []string{"any/file.go"})
	require.False(t, out.Sections[0].Entries[0].SourceChanged)
}

func TestEnrichConfigDiffWithSourceChangedRootDirMatchesSameRepo(t *testing.T) {
	full := &ComputeAppConfigDiffOutput{
		Sections: []ConfigDiffSection{
			{
				Name: "Components",
				Entries: []ConfigDiffEntry{
					{Op: "change", Name: "alb"},
					{Op: "change", Name: "pulumi"},
				},
			},
		},
	}

	sources := []componentSource{
		{Name: "alb", Repo: "acme/app", Directory: "."},
		{Name: "pulumi", Repo: "acme/app", Directory: "components/pulumi"},
	}
	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", []string{"inputs/dns/domain.toml"})
	require.True(t, out.Sections[0].Entries[0].SourceChanged)
	require.False(t, out.Sections[0].Entries[1].SourceChanged)
}
