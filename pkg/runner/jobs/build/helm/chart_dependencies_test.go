package helm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/common"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

func TestChartRefreshRemoteDependenciesInMemory(t *testing.T) {
	for _, oci := range []bool{false, true} {
		t.Run(fmt.Sprintf("oci=%t", oci), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			repository := chartRepositoryFixture(t, oci, false)
			metadata := fmt.Sprintf("apiVersion: v2\nname: parent\nversion: 1.0.0\ndependencies:\n- name: acme\n  alias: first\n  version: ^1.0.0\n  repository: %q\n- name: acme\n  alias: second\n  version: ^1.0.0\n  repository: %q\n", repository, repository)
			require.NoError(t, afero.WriteFile(fs, "Chart.yaml", []byte(metadata), 0600))
			source, err := sourcefs.From(fs, "/memory/chart")
			require.NoError(t, err)
			defer source.Close()
			h := &handler{state: initializedChartState(t, source)}
			packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), ".")
			require.NoError(t, err)
			chart, err := loader.Load(packagePath)
			require.NoError(t, err)
			require.Len(t, chart.Dependencies(), 1)
			require.Equal(t, "1.2.3+meta", chart.Dependencies()[0].Metadata.Version)
			require.Len(t, chart.Lock.Dependencies, 2)
			require.Equal(t, "1.2.3+meta", chart.Lock.Dependencies[0].Version)
			require.Equal(t, "^1.0.0", chart.Metadata.Dependencies[0].Version)
			require.Equal(t, "first", chart.Metadata.Dependencies[0].Alias)
			require.Contains(t, chart.Lock.Digest, "sha256:")
			_, err = fs.Stat("Chart.lock")
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestChartRefreshDeduplicatesRepositoryDownloads(t *testing.T) {
	data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: "apiVersion: v2\nname: acme\nversion: 1.2.3\n"})
	var indexes, archives atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/charts/index.yaml", "/other/index.yaml":
			indexes.Add(1)
			fmt.Fprint(w, "apiVersion: v1\nentries:\n  acme:\n  - apiVersion: v2\n    name: acme\n    version: 1.2.3\n    urls: [/archives/acme.tgz]\n")
		case "/archives/acme.tgz":
			archives.Add(1)
			_, _ = w.Write(data)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	chart := chartNamed("parent")
	chart.Metadata.APIVersion, chart.Metadata.Version = "v2", "1.0.0"
	chart.Raw = []*common.File{{Name: "Chart.yaml", Data: []byte("apiVersion: v2\nname: parent\nversion: 1.0.0\n")}}
	chart.Metadata.Dependencies = []*chartv2.Dependency{
		{Name: "acme", Alias: "first", Version: "^1.0.0", Repository: server.URL + "/charts"},
		{Name: "acme", Alias: "second", Version: ">=1.2.0", Repository: server.URL + "/charts/"},
		{Name: "acme", Alias: "third", Version: "1.2.3", Repository: server.URL + "/other"},
	}
	for refresh := int32(1); refresh <= 2; refresh++ {
		require.NoError(t, refreshChartDependencies(t.Context(), server.Client(), nil, chart))
		require.Equal(t, 2*refresh, indexes.Load(), "each resolved index URL is fetched once per refresh")
		require.Equal(t, refresh, archives.Load(), "different constraints/repositories resolving to one URL share an archive")
		require.Len(t, chart.Dependencies(), 1)
		require.Len(t, chart.Lock.Dependencies, 3)
		for _, dep := range chart.Lock.Dependencies {
			require.Equal(t, "1.2.3", dep.Version)
		}
		require.Equal(t, "^1.0.0", chart.Metadata.Dependencies[0].Version)
		require.Equal(t, "second", chart.Metadata.Dependencies[1].Alias)
	}
}

func TestChartRefreshPrunesOnlyObsoleteArchives(t *testing.T) {
	for _, backend := range []string{"memory", "os"} {
		t.Run(backend, func(t *testing.T) {
			var fs afero.Fs = afero.NewMemMapFs()
			root := t.TempDir()
			if backend == "os" {
				fs, _ = pinnedChartSource(t, root)
			}
			for name, data := range map[string]string{
				"chart/Chart.yaml":                               acmeChartMetadata + "dependencies:\n- name: child\n  version: 1.0.0\n  repository: file://../child\n",
				"child/Chart.yaml":                               "apiVersion: v2\nname: child\nversion: 1.0.0\n",
				"child/templates/config.yaml":                    "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: refreshed-child\n",
				"chart/charts/custom-path/Chart.yaml":            "apiVersion: v2\nname: directory-child\nversion: 2.0.0\n",
				"chart/charts/custom-path/templates/config.yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: retained-directory\n",
			} {
				require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
				require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
			}
			obsolete := chartArchiveFixture(t,
				chartArchiveEntry{name: "obsolete/Chart.yaml", data: "apiVersion: v2\nname: obsolete\nversion: 1.0.0\n"},
				chartArchiveEntry{name: "obsolete/templates/config.yaml", data: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: obsolete-resource\n"},
			)
			require.NoError(t, afero.WriteFile(fs, "chart/charts/obsolete-1.0.0.tgz", obsolete, 0600))
			source, err := sourcefs.From(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			h := &handler{state: initializedChartState(t, source)}
			packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), "chart")
			require.NoError(t, err)
			chart, err := loader.Load(packagePath)
			require.NoError(t, err)
			var names []string
			for _, child := range chart.Dependencies() {
				names = append(names, child.Name())
			}
			require.ElementsMatch(t, []string{"child", "directory-child"}, names)
			inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
			require.NoError(t, err)
			names = nil
			for _, input := range inputs {
				names = append(names, input.Review.Object["metadata"].(map[string]interface{})["name"].(string))
			}
			require.ElementsMatch(t, []string{"refreshed-child", "retained-directory"}, names)
			unchanged, err := afero.ReadFile(fs, "chart/charts/obsolete-1.0.0.tgz")
			require.NoError(t, err)
			require.Equal(t, obsolete, unchanged, "pruning must not mutate the source")
		})
	}
}

func TestChartRefreshFailureKeepsEntireVendoredSet(t *testing.T) {
	fs := afero.NewMemMapFs()
	metadata := fmt.Sprintf("apiVersion: v2\nname: parent\nversion: 1.0.0\ndependencies:\n- name: acme\n  version: ^1.0.0\n  repository: %q\n- name: child\n  version: 0.1.0\n  repository: file://../missing\n", chartRepositoryFixture(t, false, false))
	for name, data := range map[string]string{
		"chart/Chart.yaml":              metadata,
		"chart/charts/acme/Chart.yaml":  "apiVersion: v2\nname: acme\nversion: 1.0.0\n",
		"chart/charts/child/Chart.yaml": "apiVersion: v2\nname: child\nversion: 0.1.0\n",
	} {
		require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
		require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
	}
	obsolete := chartArchiveFixture(t, chartArchiveEntry{name: "obsolete/Chart.yaml", data: "apiVersion: v2\nname: obsolete\nversion: 1.0.0\n"})
	require.NoError(t, afero.WriteFile(fs, "chart/charts/obsolete-1.0.0.tgz", obsolete, 0600))
	source, err := sourcefs.From(fs, "/memory/repo")
	require.NoError(t, err)
	defer source.Close()
	h := &handler{state: initializedChartState(t, source)}
	packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), "chart")
	require.NoError(t, err)
	chart, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, chart.Dependencies(), 3)
	require.Nil(t, chart.Lock)
	var names []string
	for _, dep := range chart.Dependencies() {
		names = append(names, dep.Name())
		if dep.Name() == "acme" {
			require.Equal(t, "1.0.0", dep.Metadata.Version, "a failed refresh must not partially update dependencies")
		}
	}
	require.ElementsMatch(t, []string{"acme", "child", "obsolete"}, names)
}

func TestV1ChartLocalRefreshKeepsLegacyLock(t *testing.T) {
	fs, source := pinnedChartSource(t, t.TempDir())
	for name, data := range map[string]string{
		"chart/Chart.yaml":        "apiVersion: v1\nname: parent\nversion: 1.0.0\n",
		"chart/requirements.yaml": "dependencies:\n- name: child\n  version: ^0.1.0\n  repository: file://../child\n",
		"child/Chart.yaml":        "apiVersion: v2\nname: child\nversion: 0.1.2\n",
	} {
		require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
		require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
	}
	h := &handler{state: initializedChartState(t, source)}
	packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), "chart")
	require.NoError(t, err)
	chart, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Equal(t, "0.1.2", chart.Lock.Dependencies[0].Version)
	require.Equal(t, "^0.1.0", chart.Metadata.Dependencies[0].Version)
	var legacyLock bool
	for _, file := range chart.Files {
		legacyLock = legacyLock || file.Name == "requirements.lock"
	}
	require.True(t, legacyLock)
	_, err = fs.Stat("chart/requirements.lock")
	require.True(t, os.IsNotExist(err))
}

func TestFileDependencySymlinkWithinRepository(t *testing.T) {
	root := t.TempDir()
	fs, source := pinnedChartSource(t, root)
	for name, data := range map[string]string{
		"chart/Chart.yaml": acmeChartMetadata + "dependencies:\n- name: child\n  version: 0.1.0\n  repository: file://../linked-child\n",
		"child/Chart.yaml": "apiVersion: v2\nname: child\nversion: 0.1.0\n",
	} {
		require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
		require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
	}
	require.NoError(t, os.Symlink("child", filepath.Join(root, "linked-child")))
	h := &handler{state: initializedChartState(t, source)}
	packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), "chart")
	require.NoError(t, err)
	chart, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, chart.Dependencies(), 1)
	require.Equal(t, "child", chart.Dependencies()[0].Name())
	require.Equal(t, "0.1.0", chart.Lock.Dependencies[0].Version)
}
