package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/rootfs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const acmeChartMetadata = "apiVersion: v2\nname: acme\nversion: 1.0.0\n"

func TestHelmBuildRejectsUnsafeFilesystemWrappers(t *testing.T) {
	for name, fs := range map[string]afero.Fs{
		"base path": afero.NewBasePathFs(afero.NewOsFs(), t.TempDir()),
		"readonly":  afero.NewReadOnlyFs(afero.NewOsFs()),
		"overlay":   afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs()),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := New(HandlerParams{FS: fs})
			require.ErrorContains(t, err, "unsupported helm build filesystem")
		})
	}
}

type chartArchiveEntry struct {
	name, data, link string
	kind             byte
}

func chartArchiveFixture(t *testing.T, entries ...chartArchiveEntry) []byte {
	t.Helper()
	var data bytes.Buffer
	z := gzip.NewWriter(&data)
	tarWriter := tar.NewWriter(z)
	for _, entry := range entries {
		kind := entry.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		require.NoError(t, tarWriter.WriteHeader(&tar.Header{
			Name: entry.name, Typeflag: kind, Linkname: entry.link,
			Size: int64(len(entry.data)), Mode: 0644, ModTime: time.Unix(1700000000, 0),
		}))
		_, err := tarWriter.Write([]byte(entry.data))
		require.NoError(t, err)
	}
	require.NoError(t, tarWriter.Close())
	require.NoError(t, z.Close())
	return data.Bytes()
}

func pinnedChartSource(t *testing.T, root string) (*rootfs.FS, *sourcefs.FS) {
	t.Helper()
	fs, err := rootfs.Open(root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fs.Close()) })
	source, err := sourcefs.From(fs, root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	return fs, source
}

func TestChartMetadataDataIgnoreAndSchemaSymlinks(t *testing.T) {
	files := map[string]string{
		"Chart.yaml":            acmeChartMetadata,
		"values.yaml":           "configName: in-chart\n",
		"values.schema.json":    `{"type":"object"}`,
		".helmignore":           "excluded.txt\n",
		"templates/config.yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Values.configName }}\n",
	}
	for name, data := range files {
		for _, inside := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/inside=%t", name, inside), func(t *testing.T) {
				root := t.TempDir()
				chartDir := filepath.Join(root, "chart")
				for path, contents := range files {
					path = filepath.Join(chartDir, path)
					require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
					require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
				}
				require.NoError(t, os.WriteFile(filepath.Join(chartDir, "excluded.txt"), []byte("not packaged"), 0600))
				target := filepath.Join(root, "payload")
				if inside {
					target = filepath.Join(chartDir, "payload")
				}
				require.NoError(t, os.WriteFile(target, []byte(data), 0600))
				link := filepath.Join(chartDir, name)
				require.NoError(t, os.Remove(link))
				require.NoError(t, os.Symlink(target, link))
				_, source := pinnedChartSource(t, root)
				h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
				_, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
				if !inside {
					require.Error(t, err, "chart reads must not escape even to another file in the repository")
					return
				}
				require.NoError(t, err)
				inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
				require.NoError(t, err)
				require.Len(t, inputs, 1)
				require.Equal(t, "in-chart", inputs[0].Review.Object["metadata"].(map[string]interface{})["name"])
				for _, file := range h.state.chart.Files {
					require.NotEqual(t, "excluded.txt", file.Name)
				}
			})
		}
	}
}

func TestChartRootSymlinkContainment(t *testing.T) {
	for _, inside := range []bool{false, true} {
		t.Run(fmt.Sprintf("inside=%t", inside), func(t *testing.T) {
			root, target := t.TempDir(), t.TempDir()
			if inside {
				target = filepath.Join(root, "real-chart")
				require.NoError(t, os.Mkdir(target, 0700))
			}
			require.NoError(t, os.WriteFile(filepath.Join(target, "Chart.yaml"), []byte(acmeChartMetadata), 0600))
			require.NoError(t, os.Symlink(target, filepath.Join(root, "chart")))
			_, source := pinnedChartSource(t, root)
			chart, err := loadChart(source, "chart")
			if !inside {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "acme", chart.Name())
		})
	}
}

func TestChartArchiveRejectsMaliciousEntriesBeforeMaterialization(t *testing.T) {
	nested := chartArchiveFixture(t,
		chartArchiveEntry{name: "child/Chart.yaml", data: "apiVersion: v2\nname: child\nversion: 0.1.0\n"},
		chartArchiveEntry{name: "child/link", kind: tar.TypeSymlink, link: "../../canary"},
	)
	cases := map[string]chartArchiveEntry{
		"parent":         {name: "acme/../../canary", data: "overwritten"},
		"inner parent":   {name: "acme/templates/../canary", data: "overwritten"},
		"absolute":       {name: "/acme/canary", data: "overwritten"},
		"windows":        {name: `acme\..\canary`, data: "overwritten"},
		"drive":          {name: "acme/C:/canary", data: "overwritten"},
		"directory":      {name: "acme/../../canary/", kind: tar.TypeDir},
		"base traversal": {name: "../canary", data: "overwritten"},
		"symlink":        {name: "acme/link", kind: tar.TypeSymlink, link: "../../canary"},
		"hardlink":       {name: "acme/link", kind: tar.TypeLink, link: "../../canary"},
		"fifo":           {name: "acme/fifo", kind: tar.TypeFifo},
		"duplicate":      {name: "acme/Chart.yaml", data: acmeChartMetadata},
		"second base":    {name: "other/file", data: "overwritten"},
		"nested archive": {name: "acme/charts/child.tgz", data: string(nested)},
	}
	for name, entry := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "job")
			require.NoError(t, os.Mkdir(root, 0700))
			canary := filepath.Join(parent, "canary")
			require.NoError(t, os.WriteFile(canary, []byte("untouched"), 0600))
			fs, _ := pinnedChartSource(t, root)
			data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: acmeChartMetadata}, entry)
			_, err := materializeChartArchive(fs, bytes.NewReader(data))
			require.Error(t, err)
			entries, err := afero.ReadDir(fs, ".")
			require.NoError(t, err)
			require.Empty(t, entries, "invalid archive must not publish even its valid prefix")
			got, err := os.ReadFile(canary)
			require.NoError(t, err)
			require.Equal(t, "untouched", string(got))
		})
	}
}

func TestChartArchiveMaterializationPreservesHelmSemantics(t *testing.T) {
	child := chartArchiveFixture(t,
		chartArchiveEntry{name: "child/Chart.yaml", data: "apiVersion: v2\nname: child\nversion: 0.1.0\nannotations:\n  example.com/metadata: retained\n"},
		chartArchiveEntry{name: "child/README.md", data: "child README"},
	)
	data := chartArchiveFixture(t,
		chartArchiveEntry{name: "acme/Chart.yaml", data: acmeChartMetadata},
		chartArchiveEntry{name: "acme/.helmignore", data: "ignored.txt\n"},
		chartArchiveEntry{name: "acme/ignored.txt", data: "not packaged"},
		chartArchiveEntry{name: "acme/values.yaml", data: "name: from-archive\n"},
		chartArchiveEntry{name: "acme/templates/config.yaml", data: "\xef\xbb\xbfapiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Values.name }}\n"},
		chartArchiveEntry{name: "acme/charts/child.tgz", data: string(child)},
	)
	fs, source := pinnedChartSource(t, t.TempDir())
	dir, err := materializeChartArchive(fs, bytes.NewReader(data))
	require.NoError(t, err)
	want, err := loader.Load(filepath.Join(fs.Root(), dir))
	require.NoError(t, err)
	got, err := loadChart(source, dir)
	require.NoError(t, err)
	require.Equal(t, want, got)
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), dir, t.TempDir())
	require.NoError(t, err)
	packaged, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Equal(t, "retained", packaged.Dependencies()[0].Metadata.Annotations["example.com/metadata"])
	require.True(t, packaged.ModTime.Equal(time.Unix(1700000000, 0)))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(dir, "templates/config.yaml"), []byte("{{ invalid"), 0600))
	inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.Equal(t, "from-archive", inputs[0].Review.Object["metadata"].(map[string]interface{})["name"])
}

func TestFileDependenciesCannotEscapeSourceEvenWhenVendored(t *testing.T) {
	for _, mode := range []string{"parent", "absolute", "symlink", "dependency name"} {
		t.Run(mode, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "repo")
			chartDir := filepath.Join(root, "chart")
			outside := filepath.Join(parent, "neighbor")
			require.NoError(t, os.MkdirAll(filepath.Join(chartDir, "charts/child"), 0700))
			require.NoError(t, os.MkdirAll(outside, 0700))
			childMetadata := "apiVersion: v2\nname: child\nversion: 0.1.0\n"
			require.NoError(t, os.WriteFile(filepath.Join(outside, "Chart.yaml"), []byte(childMetadata), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(outside, "canary"), []byte("untouched"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(chartDir, "charts/child/Chart.yaml"), []byte(childMetadata), 0600))
			depName, repository := "child", "file://../../neighbor"
			switch mode {
			case "absolute":
				repository = "file://" + outside
			case "symlink":
				require.NoError(t, os.Symlink(outside, filepath.Join(root, "linked")))
				repository = "file://../linked"
			case "dependency name":
				depName, repository = "../../neighbor", ""
			}
			metadata := acmeChartMetadata + fmt.Sprintf("dependencies:\n- name: %q\n  version: 0.1.0\n  repository: %q\n", depName, repository)
			require.NoError(t, os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(metadata), 0600))
			_, source := pinnedChartSource(t, root)
			h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
			dst := t.TempDir()
			_, err := h.loadAndPackageChart(zap.NewNop(), chartDir, dst)
			require.Error(t, err)
			_, err = os.Stat(filepath.Join(dst, defaultChartPackageFilename))
			require.True(t, os.IsNotExist(err))
			canary, err := os.ReadFile(filepath.Join(outside, "canary"))
			require.NoError(t, err)
			require.Equal(t, "untouched", string(canary))
		})
	}
}

func TestConfinedDependencyRefreshMemoryAndOS(t *testing.T) {
	for _, backend := range []string{"memory", "os"} {
		for _, local := range []string{"../child", "charts/child"} {
			t.Run(backend+"/"+local, func(t *testing.T) {
				root := t.TempDir()
				var fs afero.Fs = afero.NewMemMapFs()
				if backend == "os" {
					pinned, _ := pinnedChartSource(t, root)
					fs = pinned
				}
				metadata := acmeChartMetadata + fmt.Sprintf("dependencies:\n- name: child\n  version: 0.1.0\n  repository: file://%s\n", local)
				for name, data := range map[string]string{
					"chart/Chart.yaml":                                     metadata,
					filepath.Join("chart", local, "Chart.yaml"):            "apiVersion: v2\nname: child\nversion: 0.1.0\n",
					filepath.Join("chart", local, "templates/config.yaml"): "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: refreshed-child\n",
				} {
					require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
					require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
				}
				source, err := sourcefs.From(fs, root)
				require.NoError(t, err)
				defer source.Close()
				h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
				packagePath, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
				require.NoError(t, err)
				packaged, err := loader.Load(packagePath)
				require.NoError(t, err)
				require.Len(t, packaged.Dependencies(), 1)
				require.NotNil(t, packaged.Lock)
				require.Equal(t, "file://"+local, packaged.Metadata.Dependencies[0].Repository)
				_, err = fs.Stat("chart/Chart.lock")
				require.True(t, os.IsNotExist(err), "refresh must not mutate source")
				inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
				require.NoError(t, err)
				require.Len(t, inputs, 1)
				require.Equal(t, "refreshed-child", inputs[0].Review.Object["metadata"].(map[string]interface{})["name"])
			})
		}
	}
}

func TestPackagePublicationDoesNotFollowOutputSymlink(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "Chart.yaml"), []byte(acmeChartMetadata), 0600))
	_, source := pinnedChartSource(t, root)
	dst := t.TempDir()
	canary := filepath.Join(t.TempDir(), "canary")
	require.NoError(t, os.WriteFile(canary, []byte("untouched"), 0600))
	require.NoError(t, os.Symlink(canary, filepath.Join(dst, defaultChartPackageFilename)))
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	_, err := h.loadAndPackageChart(zap.NewNop(), ".", dst)
	require.Error(t, err)
	got, err := os.ReadFile(canary)
	require.NoError(t, err)
	require.Equal(t, "untouched", string(got))
	entries, err := os.ReadDir(dst)
	require.NoError(t, err)
	require.Len(t, entries, 1, "private dependency/serializer staging must be removed on failure")
}

func TestChartSourceSurvivesDisplayPathReplacement(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "chart"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "chart/Chart.yaml"), []byte(acmeChartMetadata), 0600))
	_, source := pinnedChartSource(t, root)
	require.NoError(t, os.Rename(root, root+"-moved"))
	outside := t.TempDir()
	require.NoError(t, os.Symlink(outside, root))
	got, err := loadChart(source, filepath.Join(root, "chart"))
	require.NoError(t, err)
	require.Equal(t, "acme", got.Name())
}

func TestHelmSettingsAreJobPrivate(t *testing.T) {
	outside := t.TempDir()
	for _, name := range []string{"HELM_REPOSITORY_CONFIG", "HELM_REPOSITORY_CACHE", "HELM_REGISTRY_CONFIG", "HELM_CONTENT_CACHE", "HELM_PLUGINS"} {
		t.Setenv(name, filepath.Join(outside, name))
	}
	fs, source := pinnedChartSource(t, t.TempDir())
	require.NoError(t, afero.WriteFile(fs, "Chart.yaml", []byte(acmeChartMetadata), 0600))
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	_, err := h.loadAndPackageChart(zap.NewNop(), ".", t.TempDir())
	require.NoError(t, err)
	entries, err := os.ReadDir(outside)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestPackagedPolicySchemaCannotReadExternalFiles(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "external.schema.json")
	canary := []byte(`{"properties":{"configName":{"const":"external-allow"}}}`)
	require.NoError(t, os.WriteFile(outside, canary, 0600))
	for backend, fs := range map[string]afero.Fs{"memory": afero.NewMemMapFs(), "os": afero.NewOsFs()} {
		t.Run(backend, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, fs.MkdirAll(root, 0700))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			for name, data := range map[string]string{
				"Chart.yaml":            acmeChartMetadata,
				"values.yaml":           "configName: external-allow\n",
				"values.schema.json":    fmt.Sprintf(`{"$ref":%q}`, filepath.ToSlash(outside)),
				"templates/config.yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Values.configName }}\n",
			} {
				name = filepath.Join(root, name)
				require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0700))
				require.NoError(t, afero.WriteFile(fs, name, []byte(data), 0600))
			}
			h, err := New(HandlerParams{FS: fs})
			require.NoError(t, err)
			h.state = &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}
			_, err = h.loadAndPackageChart(zap.NewNop(), ".", t.TempDir())
			require.NoError(t, err)
			inputs, err := h.buildPolicyInput(context.Background(), zap.NewNop())
			require.ErrorContains(t, err, "external schema resource is not allowed")
			require.Nil(t, inputs)
		})
	}
	got, err := os.ReadFile(outside)
	require.NoError(t, err)
	require.Equal(t, canary, got)
}

func TestChartArchiveRejectsUnsafeOutputComponents(t *testing.T) {
	for _, name := range []string{"../canary", ".", "..", `acme\canary`, "C:canary"} {
		t.Run(name, func(t *testing.T) {
			fs, _ := pinnedChartSource(t, t.TempDir())
			metadata := fmt.Sprintf("apiVersion: v2\nname: %q\nversion: 1.0.0\n", name)
			data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: metadata})
			_, err := materializeChartArchive(fs, bytes.NewReader(data))
			require.Error(t, err)
			entries, err := afero.ReadDir(fs, ".")
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestMissingLocalDependencyRetainsVendoredFallback(t *testing.T) {
	fs, source := pinnedChartSource(t, t.TempDir())
	require.NoError(t, fs.MkdirAll("chart/charts", 0700))
	metadata := acmeChartMetadata + "dependencies:\n- name: child\n  version: 0.1.0\n  repository: file://../missing\n"
	require.NoError(t, afero.WriteFile(fs, "chart/Chart.yaml", []byte(metadata), 0600))
	child := chartArchiveFixture(t, chartArchiveEntry{name: "child/Chart.yaml", data: "apiVersion: v2\nname: child\nversion: 0.1.0\n"})
	require.NoError(t, afero.WriteFile(fs, "chart/charts/child-0.1.0.tgz", child, 0600))
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), "chart", t.TempDir())
	require.NoError(t, err)
	packaged, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, packaged.Dependencies(), 1)
	require.Equal(t, "child", packaged.Dependencies()[0].Name())
}

func TestPackagedChartInputRetainsArchiveIgnoreSemantics(t *testing.T) {
	fs, source := pinnedChartSource(t, t.TempDir())
	metadata := acmeChartMetadata + "dependencies:\n- name: child\n  version: 0.1.0\n"
	child := chartArchiveFixture(t, chartArchiveEntry{name: "child/Chart.yaml", data: "apiVersion: v2\nname: child\nversion: 0.1.0\n"})
	data := chartArchiveFixture(t,
		chartArchiveEntry{name: "acme/Chart.yaml", data: metadata},
		chartArchiveEntry{name: "acme/.helmignore", data: "retained.txt\n"},
		chartArchiveEntry{name: "acme/retained.txt", data: "archives do not apply .helmignore"},
		chartArchiveEntry{name: "acme/charts/child.tgz", data: string(child)},
	)
	require.NoError(t, afero.WriteFile(fs, "input.tgz", data, 0600))
	h := &handler{state: &handlerState{source: source, cfg: &plantypes.HelmBuildPlan{}}}
	packagePath, err := h.loadAndPackageChart(zap.NewNop(), "input.tgz", t.TempDir())
	require.NoError(t, err)
	packaged, err := loader.Load(packagePath)
	require.NoError(t, err)
	require.Len(t, packaged.Dependencies(), 1)
	var retained string
	for _, file := range packaged.Files {
		if file.Name == "retained.txt" {
			retained = string(file.Data)
		}
	}
	require.Equal(t, "archives do not apply .helmignore", retained)
}

func TestHelmInitializedWorkspacePackagingAndCleanup(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	for name, fs := range map[string]afero.Fs{"default": nil, "memory": afero.NewMemMapFs(), "os": afero.NewOsFs()} {
		t.Run(name, func(t *testing.T) {
			ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
			h, err := New(HandlerParams{V: validator.New(), FS: fs})
			require.NoError(t, err)
			h.state = &handlerState{
				plan: &plantypes.BuildPlan{},
				cfg:  &plantypes.HelmBuildPlan{HelmRepoConfig: &config.HelmRepoConfig{}},
			}
			execution := &models.AppRunnerJobExecution{ID: fmt.Sprintf("helm-confinement-%d", time.Now().UnixNano())}
			job := &models.AppRunnerJob{}
			t.Cleanup(func() { require.NoError(t, h.Cleanup(ctx, job, execution)) })
			require.NoError(t, h.Initialize(ctx, job, execution))
			workspaceFS := h.state.workspace.FS()
			require.NoError(t, afero.WriteFile(workspaceFS, "Chart.yaml", []byte(acmeChartMetadata), 0600))
			packagePath, err := h.loadAndPackageChart(zap.NewNop(), ".", h.state.arch.TmpDir())
			require.NoError(t, err)
			h.state.packagePath = packagePath
			files, err := h.getSourceFiles()
			require.NoError(t, err)
			require.NoError(t, h.state.arch.Pack(ctx, zap.NewNop(), h.state.output, files))
			archivePath := h.state.arch.TmpDir()
			output := h.state.outputRoot
			require.NoError(t, h.Cleanup(ctx, job, execution))
			_, err = output.Stat(".")
			require.Error(t, err, "owned output handle must be closed")
			_, err = workspaceFS.Stat(".")
			require.Error(t, err, "workspace must be removed and its OS handle closed")
			_, err = os.Stat(archivePath)
			require.True(t, os.IsNotExist(err))
		})
	}
}
