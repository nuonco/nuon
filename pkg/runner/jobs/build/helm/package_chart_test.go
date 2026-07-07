package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

func chartNamed(name string) *chart.Chart {
	return &chart.Chart{Metadata: &chart.Metadata{Name: name}}
}

func TestVerifyDependenciesPresentAllVendored(t *testing.T) {
	root := chartNamed("root")
	root.AddDependency(chartNamed("a"), chartNamed("b"))

	declared := []*chart.Dependency{
		{Name: "a", Repository: "http://repo"},
		{Name: "b", Repository: "http://repo"},
	}

	require.NoError(t, verifyDependenciesPresent(declared, root))
}

func TestVerifyDependenciesPresentMissing(t *testing.T) {
	root := chartNamed("root")
	root.AddDependency(chartNamed("a"))

	declared := []*chart.Dependency{
		{Name: "a", Repository: "http://repo"},
		{Name: "crds", Repository: "http://repo"},
	}

	err := verifyDependenciesPresent(declared, root)
	require.Error(t, err)
	require.Contains(t, err.Error(), "crds")
	require.NotContains(t, err.Error(), "\"a\"")
}

func TestVerifyDependenciesPresentAlias(t *testing.T) {
	root := chartNamed("root")
	root.AddDependency(chartNamed("aliased"))

	declared := []*chart.Dependency{
		{Name: "original", Alias: "aliased", Repository: "http://repo"},
	}

	require.NoError(t, verifyDependenciesPresent(declared, root))
}

func TestRepoNameForURLStableAndUnique(t *testing.T) {
	a := repoNameForURL("http://repo-a")
	b := repoNameForURL("http://repo-b")

	require.Equal(t, a, repoNameForURL("http://repo-a"), "name must be stable for a URL")
	require.NotEqual(t, a, b, "distinct URLs must not collide")
	require.Contains(t, a, "dep-repo-")
}
