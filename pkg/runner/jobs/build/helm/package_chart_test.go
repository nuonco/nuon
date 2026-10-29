package helm

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func TestLoadAndPackageChartInputFS(t *testing.T) {
	for name, fs := range map[string]afero.Fs{
		"memory": afero.NewMemMapFs(),
		"os":     afero.NewOsFs(),
	} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			chartDir := filepath.Join(root, "chart")
			modTime := time.Unix(1700000000, 0)
			for path, data := range map[string]string{
				"Chart.yaml":               "\xef\xbb\xbfapiVersion: v2\nname: acme\nversion: 1.2.3\nannotations:\n  example.com/source: injected\n",
				"values.yaml":              "configName: from-input\n",
				"values.schema.json":       `{"type":"object"}`,
				"templates/config.yaml":    "\xef\xbb\xbfapiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Values.configName }}\n",
				"templates/ignored.yaml":   "{{ invalid template",
				"templates/.hidden.yaml":   "{{ invalid template",
				"excluded/Chart.yaml":      "invalid chart metadata",
				"secret.txt":               "not packaged",
				".helmignore":              "\xef\xbb\xbftemplates/ignored.yaml\nexcluded/\nsecret.txt\n",
				"charts/child/Chart.yaml":  "apiVersion: v2\nname: child\nversion: 0.1.0\n",
				"charts/child/values.yaml": "enabled: true\n",
				"charts/child/README.md":   "vendored dependency",
			} {
				path = filepath.Join(chartDir, path)
				require.NoError(t, fs.MkdirAll(filepath.Dir(path), 0755))
				require.NoError(t, afero.WriteFile(fs, path, []byte(data), 0644))
				require.NoError(t, fs.Chtimes(path, modTime, modTime))
			}
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			if name == "memory" {
				_, err := os.Stat(chartDir)
				require.True(t, os.IsNotExist(err), "memory input must not be mirrored to disk")
			} else {
				want, err := loader.Load(chartDir)
				require.NoError(t, err)
				got, err := loadChart(source, chartDir)
				require.NoError(t, err)
				require.Equal(t, want, got, "rooted loader must retain Helm's directory semantics")
			}

			h, err := New(HandlerParams{FS: fs})
			require.NoError(t, err)
			h.state = &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}
			packagePath, err := h.loadAndPackageChart(zap.NewNop(), chartDir, t.TempDir())
			require.NoError(t, err)
			packaged, err := loader.Load(packagePath)
			require.NoError(t, err)
			require.Equal(t, "acme", packaged.Name())
			require.Equal(t, "1.2.3", packaged.Metadata.Version)
			require.Equal(t, "injected", packaged.Metadata.Annotations["example.com/source"])
			require.True(t, packaged.ModTime.Equal(modTime))
			require.JSONEq(t, `{"type":"object"}`, string(packaged.Schema))
			require.Len(t, packaged.Dependencies(), 1)
			require.Equal(t, "child", packaged.Dependencies()[0].Name())
			require.Len(t, packaged.Templates, 1)
			require.Equal(t, "templates/config.yaml", packaged.Templates[0].Name)
			require.True(t, packaged.Templates[0].ModTime.Equal(modTime))
			for _, file := range packaged.Files {
				require.NotEqual(t, "secret.txt", file.Name)
				require.NotEqual(t, "excluded/Chart.yaml", file.Name)
			}

			// Policy input must describe the packaged chart, not a second source read.
			require.NoError(t, afero.WriteFile(fs, filepath.Join(chartDir, "templates/config.yaml"), []byte("{{ invalid template"), 0644))
			inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
			require.NoError(t, err)
			require.Len(t, inputs, 1)
			require.Equal(t, "ConfigMap", inputs[0].Review.Kind.Kind)
			require.Equal(t, "from-input", inputs[0].Review.Object["metadata"].(map[string]interface{})["name"])
		})
	}
}

func TestLoadChartSymlinkContainment(t *testing.T) {
	for _, target := range []string{"config.yaml", "../outside.yaml", "."} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			chartDir := filepath.Join(root, "chart")
			require.NoError(t, os.MkdirAll(chartDir, 0755))
			require.NoError(t, os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: acme\nversion: 1.0.0\n"), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(chartDir, "config.yaml"), []byte("inside"), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(root, "outside.yaml"), []byte("outside"), 0644))
			require.NoError(t, os.Symlink(target, filepath.Join(chartDir, "linked.yaml")))
			source, err := sourcefs.New(afero.NewOsFs(), root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			got, err := loadChart(source, chartDir)
			if target != "config.yaml" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var linked string
			for _, file := range got.Files {
				if file.Name == "linked.yaml" {
					linked = string(file.Data)
				}
			}
			require.Equal(t, "inside", linked)
		})
	}
}

func TestLoadAndPackageChartRefreshesLocalDependency(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HELM_REPOSITORY_CONFIG", filepath.Join(root, "repositories.yaml"))
	t.Setenv("HELM_REPOSITORY_CACHE", filepath.Join(root, "repository-cache"))
	t.Setenv("HELM_REGISTRY_CONFIG", filepath.Join(root, "registry.json"))
	t.Setenv("HELM_CACHE_HOME", filepath.Join(root, "cache"))
	chartDir := filepath.Join(root, "chart")
	require.NoError(t, os.MkdirAll(chartDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "child/templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte("apiVersion: v2\nname: acme\nversion: 1.0.0\ndependencies:\n- name: child\n  version: 0.1.0\n  repository: file://../child\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "child/Chart.yaml"), []byte("apiVersion: v2\nname: child\nversion: 0.1.0\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "child/templates/config.yaml"), []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: refreshed-child\n"), 0644))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), chartDir, t.TempDir())
	require.NoError(t, err)
	packaged, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, packaged.Dependencies(), 1)
	require.Equal(t, "child", packaged.Dependencies()[0].Name())
	require.NotNil(t, packaged.Lock)
	inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.Equal(t, "refreshed-child", inputs[0].Review.Object["metadata"].(map[string]interface{})["name"])
}

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
