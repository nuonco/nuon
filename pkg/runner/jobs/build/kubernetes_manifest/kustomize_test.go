package kubernetes_manifest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func kustomizeFixture(t *testing.T, fs afero.Fs, files map[string]string) {
	t.Helper()
	for name, data := range files {
		require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
		require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
	}
}

func kustomizeHandler(t *testing.T, fs afero.Fs, root string, config *plantypes.KustomizeBuildConfig) *handler {
	t.Helper()
	source, err := sourcefs.From(fs, root)
	require.NoError(t, err)
	return &handler{state: &handlerState{
		source: source,
		cfg:    &plantypes.KubernetesManifestBuildPlan{KustomizeConfig: config},
	}}
}

func TestKustomizeMemoryBuild(t *testing.T) {
	fs := afero.NewMemMapFs()
	kustomizeFixture(t, fs, map[string]string{
		"base/kustomization.yaml": `resources: [deployment.yaml]
generators: [settings.yaml]
generatorOptions:
  disableNameSuffixHash: true
`,
		"base/settings.yaml": `apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: settings
name: settings
envs: [config.env]
literals: [URL=https://example.com/api]
options:
  disableNameSuffixHash: true
`,
		"base/deployment.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 1
  selector:
    matchLabels: {app: api}
  template:
    metadata:
      labels: {app: api}
    spec:
      containers:
      - name: api
        image: example/api:v1
`,
		"base/config.env": "MODE=memory\n",
		"components/labels/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
labels:
- pairs: {owner: job}
`,
		"overlays/prod/kustomization.yaml": `resources: [../../base]
components: [../../components/labels]
namePrefix: prod-
replicas:
- name: api
  count: 3
images:
- name: example/api
  newTag: v2
transformers:
- |
  apiVersion: builtin
  kind: AnnotationsTransformer
  metadata:
    name: annotate
  annotations: {environment: test}
  fieldSpecs:
  - path: metadata/annotations
    create: true
`,
	})
	h := kustomizeHandler(t, fs, "/job", nil)
	data, err := h.buildKustomization("overlays/prod")
	require.NoError(t, err)
	objects := make(map[string]map[string]any)
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var object map[string]any
		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		objects[object["kind"].(string)] = object
	}
	require.Len(t, objects, 2)
	deployment := objects["Deployment"]
	require.Equal(t, "prod-api", deployment["metadata"].(map[string]any)["name"])
	require.Equal(t, map[string]any{"owner": "job"}, deployment["metadata"].(map[string]any)["labels"])
	require.Equal(t, map[string]any{"environment": "test"}, deployment["metadata"].(map[string]any)["annotations"])
	spec := deployment["spec"].(map[string]any)
	require.Equal(t, 3, spec["replicas"])
	containers := spec["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)
	require.Equal(t, "example/api:v2", containers[0].(map[string]any)["image"])
	require.Equal(t, "prod-settings", objects["ConfigMap"]["metadata"].(map[string]any)["name"])
	require.Equal(t, map[string]any{"MODE": "memory", "URL": "https://example.com/api"}, objects["ConfigMap"]["data"])
}

func TestKustomizeVersionedLocalNames(t *testing.T) {
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "executed")
	t.Setenv("NUON_TEST_EXEC_MARKER", marker)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	require.NoError(t, os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\nprintf ran > \"$NUON_TEST_EXEC_MARKER\"\nexit 1\n"), 0700))
	for _, backend := range []string{"memory", "rooted"} {
		t.Run(backend, func(t *testing.T) {
			var fs afero.Fs = afero.NewMemMapFs()
			path := "/job"
			if backend == "rooted" {
				root, err := rootfs.Open(t.TempDir())
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, root.Close()) })
				fs, path = root, root.Root()
			}
			kustomizeFixture(t, fs, map[string]string{
				"kustomization.yaml": `resources: [service@1.2.3, "app:v1.yaml", "1svc@1.2.3/x"]
components: [labels@1.2.3]
configMapGenerator:
- name: settings
  files: [config@1.2.3.txt]
generatorOptions:
  disableNameSuffixHash: true
`,
				"service@1.2.3/kustomization.yaml": "resources: [resource.yaml]\n",
				"service@1.2.3/resource.yaml":      "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: versioned-resource}\n",
				"app:v1.yaml":                      "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: colon-resource}\n",
				"1svc@1.2.3/x/kustomization.yaml":  "resources: [resource.yaml]\n",
				"1svc@1.2.3/x/resource.yaml":       "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: numeric-resource}\n",
				"labels@1.2.3/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
labels:
- pairs: {version: local}
`,
				"config@1.2.3.txt": "versioned generator",
			})
			h := kustomizeHandler(t, fs, path, nil)
			data, err := h.buildKustomization(".")
			require.NoError(t, err)
			require.Contains(t, string(data), "name: versioned-resource")
			require.Contains(t, string(data), "name: colon-resource")
			require.Contains(t, string(data), "name: numeric-resource")
			require.Contains(t, string(data), "version: local")
			require.Contains(t, string(data), "config@1.2.3.txt: versioned generator")

			kustomizeFixture(t, fs, map[string]string{
				"service@1.2.3/org/repo/kustomization.yaml": "resources: [resource.yaml]\n",
				"service@1.2.3/org/repo/resource.yaml":      "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: explicit-local}\n",
			})
			for _, ref := range []string{"service@1.2.3/org/repo", "./service@1.2.3/org/repo", "missing@1.2.3"} {
				require.NoError(t, afero.WriteFile(fs, "kustomization.yaml", []byte(fmt.Sprintf("resources: [%q]\n", ref)), 0600))
				data, err := h.buildKustomization(".")
				switch ref {
				case "service@1.2.3/org/repo":
					require.ErrorContains(t, err, "requires confined fetching")
				case "./service@1.2.3/org/repo":
					require.NoError(t, err)
					require.Contains(t, string(data), "name: explicit-local")
				default:
					require.Error(t, err)
					require.NotContains(t, err.Error(), "requires confined fetching")
				}
			}
		})
	}
	_, err := os.Stat(marker)
	require.True(t, os.IsNotExist(err), "local names and rejected remote names must not invoke Git: %v", err)
}

func TestKustomizeLoadRestriction(t *testing.T) {
	for _, restrictor := range []string{"root-only", "none"} {
		t.Run(restrictor, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			kustomizeFixture(t, fs, map[string]string{
				"overlay/kustomization.yaml": "configMapGenerator:\n- name: shared\n  files: [../settings.txt]\ngeneratorOptions:\n  disableNameSuffixHash: true\n",
				"settings.txt":               "job settings",
			})
			h := kustomizeHandler(t, fs, "/job", &plantypes.KustomizeBuildConfig{LoadRestrictor: restrictor})
			_, err := h.buildKustomization("overlay")
			require.ErrorContains(t, err, "not in or below")
		})
	}
}

func TestKustomizeRejectsRemoteInputs(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: remote-canary}\n")
	}))
	t.Cleanup(server.Close)
	for name, config := range map[string]string{
		"openapi":         "openapi: {path: %q}\n",
		"crds":            "crds: [%q]\n",
		"configurations":  "configurations: [%q]\n",
		"patches":         "patches: [{path: %q}]\n",
		"json patches":    "patchesJson6902: [{path: %q}]\n",
		"strategic patch": "patchesStrategicMerge: [%q]\n",
		"replacements":    "replacements: [{path: %q}]\n",
		"configmap files": "configMapGenerator: [{name: settings, files: [%q]}]\n",
		"configmap envs":  "configMapGenerator: [{name: settings, envs: [%q]}]\n",
		"secret files":    "secretGenerator: [{name: settings, files: [%q]}]\n",
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			kustomizeFixture(t, fs, map[string]string{"kustomization.yaml": fmt.Sprintf(config, server.URL+"/input")})
			_, err := kustomizeHandler(t, fs, "/job", nil).buildKustomization(".")
			require.ErrorContains(t, err, "requires confined fetching")
		})
	}
	for kind, inputs := range map[string]string{
		"ConfigMapGenerator":             "files: [%q]\n",
		"SecretGenerator":                "envs: [%q]\n",
		"PatchTransformer":               "path: %q\n",
		"PatchJson6902Transformer":       "path: %q\n",
		"PatchStrategicMergeTransformer": "paths: [%q]\n",
		"ReplacementTransformer":         "replacements: [{path: %q}]\n",
		"ValueAddTransformer":            "targetFilePath: %q\n",
	} {
		for _, inline := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/inline=%t", kind, inline), func(t *testing.T) {
				plugin := "apiVersion: builtin\nkind: " + kind + "\nmetadata: {name: remote}\n" + fmt.Sprintf(inputs, server.URL+"/input")
				fs := afero.NewMemMapFs()
				config := "transformers: [plugin.yaml]\n"
				if inline {
					config = fmt.Sprintf("transformers: [%q]\n", plugin)
				}
				kustomizeFixture(t, fs, map[string]string{"kustomization.yaml": config, "plugin.yaml": plugin})
				_, err := kustomizeHandler(t, fs, "/job", nil).buildKustomization(".")
				require.ErrorContains(t, err, "requires confined fetching")
			})
		}
	}
	require.Zero(t, requests.Load(), "all dependency reads must stay on the source filesystem")
}

func TestKustomizeRejectsSynthesizedPluginPaths(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, "remote-canary")
	}))
	t.Cleanup(server.Close)
	for _, field := range []string{"generators", "transformers", "validators"} {
		for _, restrictor := range []string{"root-only", "none"} {
			for _, linkedConfig := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/linked=%t", field, restrictor, linkedConfig), func(t *testing.T) {
					root, err := rootfs.Open(t.TempDir())
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, root.Close()) })
					kustomizeFixture(t, root, map[string]string{
						"kustomization.yaml":  fmt.Sprintf("%s: [plugins]\n", field),
						"plugins/config.yaml": "apiVersion: builtin\nkind: ConfigMapGenerator\nmetadata: {name: settings}\nname: settings\nfiles: [input]\n",
						"plugins/kustomization.yaml": fmt.Sprintf(`resources: [config.yaml]
patches:
- target: {kind: ConfigMapGenerator}
  patch: |-
    - op: replace
      path: /files/0
      value: %s
`, server.URL+"/input"),
					})
					if linkedConfig {
						require.NoError(t, root.Mkdir("config", 0700))
						require.NoError(t, root.Rename("kustomization.yaml", "config/hidden"))
						require.NoError(t, root.SymlinkIfPossible("config/hidden", "kustomization.yaml"))
					}
					_, err = kustomizeHandler(t, root, root.Root(), &plantypes.KustomizeBuildConfig{LoadRestrictor: restrictor}).buildKustomization(".")
					require.ErrorContains(t, err, "directory plugin configuration")
				})
			}
		}
	}
	require.Zero(t, requests.Load(), "rendered plugin paths must not escape source validation")
}

func TestKustomizeRejectsUnconfinedFeatures(t *testing.T) {
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "executed")
	t.Setenv("NUON_TEST_EXEC_MARKER", marker)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, executable := range []string{"git", "helm"} {
		require.NoError(t, os.WriteFile(filepath.Join(bin, executable), []byte("#!/bin/sh\nprintf ran > \"$NUON_TEST_EXEC_MARKER\"\nexit 1\n"), 0700))
	}
	for field, ref := range map[string]string{
		"resources":    "https://example.com/acme/base",
		"bases":        "github.com/acme/base",
		"components":   "file:///tmp/base",
		"generators":   "git@example.com:acme/base",
		"transformers": "git::https://example.com/acme/base",
		"validators":   "SSH://git@example.com/acme/base",
	} {
		t.Run(field, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			kustomizeFixture(t, fs, map[string]string{"kustomization.yaml": fmt.Sprintf("%s: [%q]\n", field, ref)})
			h := kustomizeHandler(t, fs, "/job", nil)
			_, err := h.buildKustomization(".")
			require.ErrorContains(t, err, "requires confined fetching")
		})
	}
	for name, config := range map[string]string{
		"helm charts": "helmCharts:\n- name: example\n  repo: https://example.com/charts\n",
		"legacy helm": "helmChartInflationGenerator:\n- chartName: example\n",
		"helm plugin": "generators:\n- |\n  apiVersion: builtin\n  kind: HelmChartInflationGenerator\n  metadata:\n    name: example\n  name: example\n",
		"exec plugin": "generators:\n- |\n  apiVersion: example.com/v1\n  kind: Example\n  metadata:\n    name: example\n",
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			kustomizeFixture(t, fs, map[string]string{"kustomization.yaml": config})
			_, err := kustomizeHandler(t, fs, "/job", nil).buildKustomization(".")
			require.Error(t, err)
		})
	}
	fs := afero.NewMemMapFs()
	kustomizeFixture(t, fs, map[string]string{"kustomization.yaml": "resources: []\n"})
	_, err := kustomizeHandler(t, fs, "/job", &plantypes.KustomizeBuildConfig{EnableHelm: true}).buildKustomization(".")
	require.ErrorContains(t, err, "requires confined subprocess execution")
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err), "git/helm must not execute: %v", err)
}

func TestKustomizeRootedSource(t *testing.T) {
	for _, scenario := range []string{"contained link", "contained absolute link", "contained directory", "escaping link", "escaping directory", "escaping generator", "parent path", "absolute path", "sibling prefix", "root replacement", "linked kustomization"} {
		t.Run(scenario, func(t *testing.T) {
			parent := t.TempDir()
			job := filepath.Join(parent, "job")
			neighbor := filepath.Join(parent, "job-neighbor")
			require.NoError(t, os.Mkdir(job, 0700))
			require.NoError(t, os.Mkdir(neighbor, 0700))
			root, err := rootfs.Open(job)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, root.Close()) })
			kustomizeFixture(t, root, map[string]string{
				"kustomization.yaml": "resources: [resource.yaml]\n",
				"resource.yaml":      "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: job-source\n",
			})
			canary := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: neighbor-canary\n")
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "resource.yaml"), canary, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "kustomization.yaml"), []byte("resources: [resource.yaml]\n"), 0600))
			h := kustomizeHandler(t, root, job, &plantypes.KustomizeBuildConfig{LoadRestrictor: "none"})
			path := "."
			switch scenario {
			case "contained link":
				require.NoError(t, root.Rename("resource.yaml", "real.yaml"))
				require.NoError(t, root.SymlinkIfPossible("real.yaml", "resource.yaml"))
			case "contained absolute link":
				require.NoError(t, root.Rename("resource.yaml", "real.yaml"))
				require.NoError(t, root.SymlinkIfPossible(filepath.Join(root.Root(), "real.yaml"), "resource.yaml"))
			case "contained directory":
				require.NoError(t, root.Mkdir("real", 0700))
				require.NoError(t, root.Rename("resource.yaml", "real/resource.yaml"))
				require.NoError(t, root.SymlinkIfPossible("real", "linked"))
				require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte("resources: [linked/resource.yaml]\n"), 0600))
			case "escaping link":
				require.NoError(t, root.Remove("resource.yaml"))
				require.NoError(t, root.SymlinkIfPossible("../job-neighbor/resource.yaml", "resource.yaml"))
			case "escaping directory":
				require.NoError(t, root.SymlinkIfPossible("../job-neighbor", "outside"))
				require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte("resources: [outside/resource.yaml]\n"), 0600))
			case "escaping generator":
				require.NoError(t, root.SymlinkIfPossible("../job-neighbor/resource.yaml", "settings.txt"))
				require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte("configMapGenerator:\n- name: settings\n  files: [settings.txt]\n"), 0600))
			case "parent path":
				require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte("resources: [../job-neighbor/resource.yaml]\n"), 0600))
			case "absolute path":
				require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte(fmt.Sprintf("resources: [%q]\n", filepath.Join(neighbor, "resource.yaml"))), 0600))
			case "sibling prefix":
				path = neighbor
			case "root replacement":
				require.NoError(t, os.Rename(job, job+"-moved"))
				require.NoError(t, os.Symlink(neighbor, job))
			case "linked kustomization":
				require.NoError(t, root.Remove("kustomization.yaml"))
				require.NoError(t, afero.WriteFile(root, "hidden-config", []byte("resources: [file:///tmp/base]\n"), 0600))
				require.NoError(t, root.SymlinkIfPossible("hidden-config", "kustomization.yaml"))
				h.state.cfg.KustomizeConfig.LoadRestrictor = "root-only"
			}
			data, err := h.buildKustomization(path)
			if strings.HasPrefix(scenario, "contained ") || scenario == "root replacement" {
				require.NoError(t, err)
				require.Contains(t, string(data), "name: job-source")
			} else {
				require.Error(t, err)
			}
			require.NotContains(t, string(data), "neighbor-canary")
			unchanged, err := os.ReadFile(filepath.Join(neighbor, "resource.yaml"))
			require.NoError(t, err)
			require.Equal(t, canary, unchanged)
		})
	}
}

type replaceKustomizationFS struct {
	*rootfs.FS
	replace func()
}

func (f *replaceKustomizationFS) Open(name string) (afero.File, error) {
	if name == "kustomization.yaml" && f.replace != nil {
		replace := f.replace
		f.replace = nil
		replace()
	}
	return f.FS.Open(name)
}

func TestKustomizeValidatesReplacedConfiguration(t *testing.T) {
	root, err := rootfs.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	kustomizeFixture(t, root, map[string]string{"kustomization.yaml": "resources: []\n"})
	fs := &replaceKustomizationFS{FS: root, replace: func() {
		require.NoError(t, afero.WriteFile(root, "kustomization.yaml", []byte("resources: [file:///tmp/base]\n"), 0600))
	}}
	_, err = kustomizeHandler(t, fs, root.Root(), nil).buildKustomization(".")
	require.ErrorContains(t, err, "requires confined fetching")
}
