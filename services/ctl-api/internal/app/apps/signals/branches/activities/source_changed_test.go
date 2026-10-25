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

	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", changed, nil, nil, nil)
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

func TestMarkComponentSourceChangedAddsSourceOnlyComponent(t *testing.T) {
	out := &ConfigDiffWithSourceOutput{
		Sections: []ConfigDiffSectionWithSource{
			{Name: "Inputs", Entries: []ConfigDiffEntryWithSource{{Op: "change", Name: "dns"}}},
		},
		ComponentSourceChanged: map[string]bool{},
	}
	markComponentSourceChanged(out, "application_load_balancer")

	require.True(t, out.ComponentSourceChanged["application_load_balancer"])
	require.False(t, out.ComponentSourceChanged["unchanged"])
	require.Equal(t, 1, out.Changed)

	var components *ConfigDiffSectionWithSource
	for i := range out.Sections {
		if out.Sections[i].Name == "Components" {
			components = &out.Sections[i]
		}
	}
	require.NotNil(t, components)
	require.Equal(t, "application_load_balancer", components.Entries[0].Name)
	require.True(t, components.Entries[0].SourceChanged)
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
	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", []string{"components/api/main.go"}, nil, nil, nil)
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

	out := enrichConfigDiffWithSourceChanged(full, nil, "acme/app", []string{"any/file.go"}, nil, nil, nil)
	require.False(t, out.Sections[0].Entries[0].SourceChanged)
}

func TestComponentSourceFiles(t *testing.T) {
	branch := "acme/app"
	changed := []string{"components/grafana/values.yaml", "components/grafana/templates/gcp.yaml", "docs/readme.md"}
	patches := map[string]string{
		"components/grafana/values.yaml":        "+ securityPolicy: \"\"",
		"components/grafana/templates/gcp.yaml": "+ kind: GCPBackendPolicy",
		"docs/readme.md":                        "+ unrelated",
	}

	files := componentSourceFiles(componentSource{
		Name: "grafana", Repo: "https://github.com/acme/app.git", Directory: "components/grafana",
	}, branch, changed, patches)
	require.Len(t, files, 2)
	require.ElementsMatch(t, []string{"components/grafana/values.yaml", "components/grafana/templates/gcp.yaml"},
		[]string{files[0].Path, files[1].Path})

	// Different repo: no files, even though directories/paths line up.
	require.Nil(t, componentSourceFiles(componentSource{
		Name: "grafana", Repo: "acme/other", Directory: "components/grafana",
	}, branch, changed, patches))

	// No patches available (e.g. no git diff computed): no files.
	require.Nil(t, componentSourceFiles(componentSource{
		Name: "grafana", Repo: "acme/app", Directory: "components/grafana",
	}, branch, changed, nil))
}

func TestEnrichConfigDiffWithSourceChangedAttachesSourceFiles(t *testing.T) {
	full := &ComputeAppConfigDiffOutput{
		Sections: []ConfigDiffSection{
			{
				Name:    "Components",
				Entries: []ConfigDiffEntry{{Op: "change", Name: "grafana"}},
			},
		},
	}
	sources := []componentSource{
		{Name: "grafana", Repo: "acme/app", Directory: "components/grafana"},
	}
	changed := []string{"components/grafana/values.yaml"}
	patches := map[string]string{"components/grafana/values.yaml": "+ securityPolicy: \"\""}

	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", changed, patches, nil, nil)
	require.True(t, out.Sections[0].Entries[0].SourceChanged)
	require.Len(t, out.Sections[0].Entries[0].SourceFiles, 1)
	require.Equal(t, "components/grafana/values.yaml", out.Sections[0].Entries[0].SourceFiles[0].Path)
	require.Equal(t, "+ securityPolicy: \"\"", out.Sections[0].Entries[0].SourceFiles[0].Diff)
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
	out := enrichConfigDiffWithSourceChanged(full, sources, "acme/app", []string{"inputs/dns/domain.toml"}, nil, nil, nil)
	require.True(t, out.Sections[0].Entries[0].SourceChanged)
	require.False(t, out.Sections[0].Entries[1].SourceChanged)
}
