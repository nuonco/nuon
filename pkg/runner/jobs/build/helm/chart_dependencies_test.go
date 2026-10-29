package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
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
			h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
			packagePath, err := h.loadAndPackageChart(zap.NewNop(), ".", t.TempDir())
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
	source, err := sourcefs.From(fs, "/memory/repo")
	require.NoError(t, err)
	defer source.Close()
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
	require.NoError(t, err)
	chart, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, chart.Dependencies(), 2)
	require.Nil(t, chart.Lock)
	for _, dep := range chart.Dependencies() {
		if dep.Name() == "acme" {
			require.Equal(t, "1.0.0", dep.Metadata.Version, "a failed refresh must not partially update dependencies")
		}
	}
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
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
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
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
	require.NoError(t, err)
	chart, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, chart.Dependencies(), 1)
	require.Equal(t, "child", chart.Dependencies()[0].Name())
	require.Equal(t, "0.1.0", chart.Lock.Dependencies[0].Version)
}
