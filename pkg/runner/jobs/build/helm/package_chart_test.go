package helm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func initializedChartState(t *testing.T, source *sourcefs.FS) *handlerState {
	t.Helper()
	arch := ociarchive.New()
	require.NoError(t, arch.Initialize(context.Background()))
	t.Cleanup(func() { require.NoError(t, arch.Cleanup(context.Background())) })
	output, err := sourcefs.From(arch.FS(), arch.BasePath())
	require.NoError(t, err)
	return &handlerState{source: source, arch: arch, output: output, cfg: &plantypes.HelmBuildPlan{}}
}

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
			h.state = initializedChartState(t, source)
			packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), chartDir)
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

func TestLoadChartSkipsGitMetadata(t *testing.T) {
	for _, backend := range []string{"memory", "os"} {
		for _, ignoreGit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/ignore=%t", backend, ignoreGit), func(t *testing.T) {
				var fs afero.Fs = afero.NewMemMapFs()
				root := t.TempDir()
				if backend == "os" {
					fs = afero.NewOsFs()
				}
				for name, data := range map[string]string{
					"Chart.yaml":                     acmeChartMetadata,
					".git/config":                    "not authored source",
					"charts/child/Chart.yaml":        "apiVersion: v2\nname: child\nversion: 0.1.0\n",
					"charts/child/.git":              "gitdir: outside",
					"assets/.GiT/config":             "not authored source",
					".github/workflows/example.yaml": "retained hidden content",
					"ignored.txt":                    "not packaged",
				} {
					name = filepath.Join(root, name)
					require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
					require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
				}
				ignores := "ignored.txt\n"
				if ignoreGit {
					ignores += ".git\n"
				}
				require.NoError(t, afero.WriteFile(fs, filepath.Join(root, ".helmignore"), []byte(ignores), 0600))
				if backend == "os" {
					require.NoError(t, os.Mkdir(filepath.Join(root, "linked"), 0700))
					require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(root, "linked/.git")))
				}
				source, err := sourcefs.New(fs, root)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, source.Close()) })
				got, err := loadChart(source, ".")
				require.NoError(t, err)
				require.Len(t, got.Dependencies(), 1)
				var hidden bool
				for _, file := range got.Raw {
					require.NotContains(t, strings.Split(strings.ToLower(file.Name), "/"), ".git")
					require.NotEqual(t, "ignored.txt", file.Name)
					hidden = hidden || file.Name == ".github/workflows/example.yaml"
				}
				require.True(t, hidden)
				_, err = loadChart(source, ".git")
				require.ErrorContains(t, err, "Git metadata")
				_, err = source.ReadFile(".git/config")
				require.ErrorContains(t, err, "Git metadata")
				if backend == "os" {
					require.NoError(t, os.Symlink(".git/config", filepath.Join(root, "metadata-alias")))
					got, err = loadChart(source, ".")
					require.NoError(t, err)
					for _, file := range got.Raw {
						require.NotEqual(t, "metadata-alias", file.Name)
					}
					_, err = source.ReadFile("metadata-alias")
					require.ErrorContains(t, err, "Git metadata")
				}
			})
		}
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
	h := &handler{state: initializedChartState(t, source)}
	packagePath, err := h.loadAndPackageChart(t.Context(), zap.NewNop(), chartDir)
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

func TestLoadAndPackageChartCancelsDependencyFetch(t *testing.T) {
	for _, stage := range []string{"index", "archive", "oci"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			canceled := make(chan bool, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if stage == "archive" && r.URL.Path == "/charts/index.yaml" {
					fmt.Fprint(w, "apiVersion: v1\nentries:\n  acme:\n  - apiVersion: v2\n    name: acme\n    version: 1.0.0\n    urls: [acme.tgz]\n")
					return
				}
				cancel()
				select {
				case <-r.Context().Done():
					canceled <- true
				case <-time.After(2 * time.Second):
					canceled <- false
					http.Error(w, "request was not canceled", http.StatusForbidden)
				}
			}))
			defer server.Close()
			repository := server.URL + "/charts"
			if stage == "oci" {
				repository = "oci://" + strings.TrimPrefix(server.URL, "http://") + "/charts"
			}
			fs, source := pinnedChartSource(t, t.TempDir())
			metadata := fmt.Sprintf("apiVersion: v2\nname: parent\nversion: 1.0.0\ndependencies:\n- name: acme\n  version: 1.0.0\n  repository: %q\n", repository)
			require.NoError(t, afero.WriteFile(fs, "Chart.yaml", []byte(metadata), 0600))
			h := &handler{state: initializedChartState(t, source)}
			_, err := h.loadAndPackageChart(ctx, zap.NewNop(), ".")
			require.ErrorContains(t, err, "context canceled")
			require.True(t, <-canceled, "job cancellation must abort the in-flight dependency request")
			_, err = h.state.arch.FS().Stat(defaultChartPackageFilename)
			require.True(t, os.IsNotExist(err))
		})
	}
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
