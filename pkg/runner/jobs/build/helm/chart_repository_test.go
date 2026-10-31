package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/registry"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/errs"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func chartRepositoryFixture(t *testing.T, oci, corrupt bool) string {
	t.Helper()
	versions := []string{"1.0.0", "2.0.0-beta.1", "1.2.3+meta"}
	archives := map[string][]byte{}
	manifests := map[string][]byte{}
	blobs := map[string][]byte{}
	entries := []map[string]interface{}{}
	for _, version := range versions {
		metadata := fmt.Sprintf("apiVersion: v2\nname: acme\nversion: %s\ndependencies:\n- name: child\n  version: ~0.1.0\n  repository: file://charts/child\n", version)
		data := chartArchiveFixture(t,
			chartArchiveEntry{name: "acme/Chart.yaml", data: metadata},
			chartArchiveEntry{name: "acme/templates/config.yaml", data: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: from-repository\n"},
			chartArchiveEntry{name: "acme/charts/child/Chart.yaml", data: "apiVersion: v2\nname: child\nversion: 0.1.2\n"},
		)
		archiveName := "acme-" + version + ".tgz"
		archives[archiveName] = data
		entries = append(entries, map[string]interface{}{"apiVersion": "v2", "name": "acme", "version": version, "urls": []string{"archives/" + archiveName}})
		config, err := json.Marshal(&chartv2.Metadata{Name: "acme", Version: version, APIVersion: "v2"})
		require.NoError(t, err)
		configDesc := ocispec.Descriptor{MediaType: registry.ConfigMediaType, Digest: digest.FromBytes(config), Size: int64(len(config))}
		layerDesc := ocispec.Descriptor{MediaType: registry.ChartLayerMediaType, Digest: digest.FromBytes(data), Size: int64(len(data))}
		blobs[configDesc.Digest.String()] = config
		blobs[layerDesc.Digest.String()] = data
		if corrupt {
			blobs[layerDesc.Digest.String()] = bytes.Repeat([]byte("x"), len(data))
		}
		manifest, err := json.Marshal(ocispec.Manifest{
			Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageManifest,
			Config: configDesc, Layers: []ocispec.Descriptor{layerDesc},
		})
		require.NoError(t, err)
		manifests[strings.ReplaceAll(version, "+", "_")] = manifest
		manifests[digest.FromBytes(manifest).String()] = manifest
	}
	index, err := json.Marshal(map[string]interface{}{"apiVersion": "v1", "entries": map[string]interface{}{"acme": entries}})
	require.NoError(t, err)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"token":"fixture","expires_in":3600}`)
			return
		}
		if oci && r.Header.Get("Authorization") != "Bearer fixture" {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm=%q,service="fixture",scope="repository:charts/acme:pull"`, server.URL+"/token"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var data []byte
		switch {
		case r.URL.Path == "/charts/index.yaml":
			data = index
		case strings.HasPrefix(r.URL.Path, "/charts/archives/"):
			data = archives[strings.TrimPrefix(r.URL.Path, "/charts/archives/")]
		case r.URL.Path == "/v2/charts/acme/tags/list":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"charts/acme","tags":["1.0.0","invalid","2.0.0-beta.1","1.2.3_meta"]}`)
			return
		case strings.HasPrefix(r.URL.Path, "/v2/charts/acme/manifests/"):
			data = manifests[strings.TrimPrefix(r.URL.Path, "/v2/charts/acme/manifests/")]
			w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
			w.Header().Set("Docker-Content-Digest", digest.FromBytes(data).String())
		case strings.HasPrefix(r.URL.Path, "/v2/charts/acme/blobs/"):
			data = blobs[strings.TrimPrefix(r.URL.Path, "/v2/charts/acme/blobs/")]
		}
		if data == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	if oci {
		return "oci://" + strings.TrimPrefix(server.URL, "http://") + "/charts"
	}
	return server.URL + "/charts"
}

func TestChartRepositoryIndexYAML(t *testing.T) {
	index, err := chartRepositoryIndex([]byte(`apiVersion: v1
entries:
  acme:
  - null
  - name: acme
    version: invalid
  - name: acme
    version: 1.0.0
    urls: [acme-1.0.0.tgz]
  - name: acme
    version: 1.2.3
    urls: [acme-1.2.3.tgz]
`))
	require.NoError(t, err)
	require.Len(t, index.Entries["acme"], 2)
	entry, err := index.Get("acme", "^1.0.0")
	require.NoError(t, err)
	require.Equal(t, "1.2.3", entry.Version)
	require.Equal(t, "v1", entry.APIVersion)
	for _, input := range []string{"", "entries: {}", "apiVersion: v1\napiVersion: v1"} {
		_, err := chartRepositoryIndex([]byte(input))
		require.Error(t, err)
	}
}

func TestChartRepositoryVersionsAndAnonymousOCI(t *testing.T) {
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "config.json"), []byte("invalid Docker credentials"), 0600))
	t.Setenv("DOCKER_CONFIG", outside)
	t.Setenv("HELM_REGISTRY_CONFIG", filepath.Join(outside, "config.json"))
	defaultTransport, defaultTimeout := http.DefaultTransport, http.DefaultClient.Timeout
	for _, oci := range []bool{false, true} {
		repository := chartRepositoryFixture(t, oci, false)
		client := newChartRepositoryClient()
		defer client.CloseIdleConnections()
		fetcher := newChartRepositoryFetcher(client)
		for constraint, expected := range map[string]string{
			"": "1.2.3+meta", "1.0.0": "1.0.0", "^1.0.0": "1.2.3+meta", ">=2.0.0-0": "2.0.0-beta.1",
		} {
			t.Run(fmt.Sprintf("oci=%t/%s", oci, constraint), func(t *testing.T) {
				data, err := fetcher.fetch(t.Context(), repository, "acme", constraint)
				require.NoError(t, err)
				files, err := chartArchiveFiles(bytes.NewReader(data))
				require.NoError(t, err)
				chart, err := loadChartFiles(files)
				require.NoError(t, err)
				require.Equal(t, expected, chart.Metadata.Version)
			})
		}
		_, err := fetcher.fetch(t.Context(), repository, "acme", ">=9.0.0")
		require.Error(t, err)
	}
	require.Same(t, defaultTransport, http.DefaultTransport)
	require.Equal(t, defaultTimeout, http.DefaultClient.Timeout)
	got, err := os.ReadFile(filepath.Join(outside, "config.json"))
	require.NoError(t, err)
	require.Equal(t, "invalid Docker credentials", string(got))
}

func TestChartRepositoryCacheRetriesFailedDownloads(t *testing.T) {
	for _, stage := range []string{"index", "archive"} {
		t.Run(stage, func(t *testing.T) {
			var indexes, archives atomic.Int32
			data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: acmeChartMetadata})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/index.yaml" {
					if indexes.Add(1) == 1 && stage == "index" {
						http.Error(w, "unavailable", http.StatusServiceUnavailable)
						return
					}
					fmt.Fprint(w, "apiVersion: v1\nentries:\n  acme:\n  - name: acme\n    version: 1.0.0\n    urls: [acme.tgz]\n")
					return
				}
				if archives.Add(1) == 1 && stage == "archive" {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}
				_, _ = w.Write(data)
			}))
			defer server.Close()
			fetcher := newChartRepositoryFetcher(server.Client())
			_, err := fetcher.fetch(t.Context(), server.URL, "acme", "1.0.0")
			require.Error(t, err)
			for range 2 {
				got, err := fetcher.fetch(t.Context(), server.URL, "acme", "1.0.0")
				require.NoError(t, err)
				require.Equal(t, data, got)
			}
			if stage == "index" {
				require.EqualValues(t, 2, indexes.Load())
				require.EqualValues(t, 1, archives.Load())
			} else {
				require.EqualValues(t, 1, indexes.Load())
				require.EqualValues(t, 2, archives.Load())
			}
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = fetcher.fetch(ctx, server.URL, "acme", "1.0.0")
			require.ErrorIs(t, err, context.Canceled, "cached results must not ignore cancellation")
		})
	}
}

func TestChartRepositoryCacheSharesOCILayers(t *testing.T) {
	repository := chartRepositoryFixture(t, true, false)
	upstream, err := url.Parse("http://" + strings.TrimPrefix(repository, "oci://"))
	require.NoError(t, err)
	upstream.Path = ""
	forward := httputil.NewSingleHostReverseProxy(upstream)
	var requests sync.Map
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count, _ := requests.LoadOrStore(r.URL.Path, &atomic.Int32{})
		count.(*atomic.Int32).Add(1)
		forward.ServeHTTP(w, r)
	}))
	defer server.Close()
	fetcher := newChartRepositoryFetcher(server.Client())
	var data []byte
	for _, version := range []string{"^1.0.0", "1.2.3+meta"} {
		got, err := fetcher.fetch(t.Context(), "oci://"+strings.TrimPrefix(server.URL, "http://")+"/charts", "acme", version)
		require.NoError(t, err)
		if data != nil {
			require.Equal(t, data, got)
		}
		data = got
	}
	layerDigest := digest.FromBytes(data).String()
	count, ok := requests.Load("/v2/charts/acme/blobs/" + layerDigest)
	require.True(t, ok)
	require.EqualValues(t, 1, count.(*atomic.Int32).Load(), "constraints sharing a resolved layer download it only once")
	index := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address := "oci://" + strings.TrimPrefix(server.URL, "http://") + "/charts/acme@" + layerDigest
		fmt.Fprintf(w, "apiVersion: v1\nentries:\n  acme:\n  - name: acme\n    version: 1.2.3+meta\n    urls: [%q]\n", address)
	}))
	defer index.Close()
	_, err = fetcher.fetch(t.Context(), index.URL, "acme", "1.2.3+meta")
	require.ErrorContains(t, err, "explicit HTTP(S) URL", "an HTTP index cannot reference even a cached OCI layer")
}

func TestOCIChartRejectsCorruptContent(t *testing.T) {
	repository := chartRepositoryFixture(t, true, true)
	client := newChartRepositoryClient()
	defer client.CloseIdleConnections()
	_, err := fetchRepositoryChart(context.Background(), client, repository, "acme", "1.0.0")
	require.Error(t, err)
}

func TestChartRepositoryEnvironmentProxy(t *testing.T) {
	if repository := os.Getenv("NUON_HELM_TEST_PROXY_REPOSITORY"); repository != "" {
		client := newChartRepositoryClient()
		defer client.CloseIdleConnections()
		data, err := fetchRepositoryChart(t.Context(), client, repository, "acme", "1.0.0")
		require.NoError(t, err)
		files, err := chartArchiveFiles(bytes.NewReader(data))
		require.NoError(t, err)
		chart, err := loadChartFiles(files)
		require.NoError(t, err)
		require.Equal(t, "1.0.0", chart.Metadata.Version)
		return
	}

	for _, oci := range []bool{false, true} {
		t.Run(fmt.Sprintf("oci=%t", oci), func(t *testing.T) {
			repository := chartRepositoryFixture(t, oci, false)
			upstream, err := url.Parse("http://" + strings.TrimPrefix(strings.TrimPrefix(repository, "oci://"), "http://"))
			require.NoError(t, err)
			upstream.Path = ""
			forward := httputil.NewSingleHostReverseProxy(upstream)
			forward.ModifyResponse = func(resp *http.Response) error {
				if challenge := resp.Header.Get("WWW-Authenticate"); challenge != "" {
					resp.Header.Set("WWW-Authenticate", strings.ReplaceAll(challenge, upstream.String(), "http://charts.example.invalid"))
				}
				return nil
			}
			var requests atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				forward.ServeHTTP(w, r)
			}))
			defer proxy.Close()
			scheme := "http"
			if oci {
				scheme = "oci"
			}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			// Go caches proxy environment variables on first use; use a fresh process.
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestChartRepositoryEnvironmentProxy$")
			cmd.Env = append(os.Environ(),
				"HTTP_PROXY="+proxy.URL, "http_proxy="+proxy.URL,
				"HTTPS_PROXY="+proxy.URL, "https_proxy="+proxy.URL,
				"NO_PROXY=", "no_proxy=", "REQUEST_METHOD=",
				"NUON_HELM_TEST_PROXY_REPOSITORY="+scheme+"://charts.example.invalid/charts",
			)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", output)
			require.Positive(t, requests.Load())
		})
	}
}

func TestOCIChartRetriesTransientFailures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		failures int32
		attempts int32
		wantErr  bool
	}{
		{name: "rate limited", status: http.StatusTooManyRequests, failures: 1, attempts: 2},
		{name: "unavailable", status: http.StatusServiceUnavailable, failures: 1, attempts: 2},
		{name: "forbidden", status: http.StatusForbidden, failures: 1, attempts: 1, wantErr: true},
		{name: "retry limit", status: http.StatusServiceUnavailable, failures: 100, attempts: 6, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := chartRepositoryFixture(t, true, false)
			upstream, err := url.Parse("http://" + strings.TrimPrefix(repository, "oci://"))
			require.NoError(t, err)
			upstream.Path = ""
			forward := httputil.NewSingleHostReverseProxy(upstream)
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/charts/acme/manifests/1.0.0" && r.Header.Get("Authorization") == "Bearer fixture" {
					if attempts.Add(1) <= tc.failures {
						w.WriteHeader(tc.status)
						return
					}
				}
				forward.ServeHTTP(w, r)
			}))
			defer server.Close()
			client := newChartRepositoryClient()
			defer client.CloseIdleConnections()
			data, err := fetchRepositoryChart(t.Context(), client, "oci://"+strings.TrimPrefix(server.URL, "http://")+"/charts", "acme", "1.0.0")
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				files, err := chartArchiveFiles(bytes.NewReader(data))
				require.NoError(t, err)
				chart, err := loadChartFiles(files)
				require.NoError(t, err)
				require.Equal(t, "1.0.0", chart.Metadata.Version)
			}
			require.Equal(t, tc.attempts, attempts.Load())
		})
	}
}

func TestChartRepositoryFailsClosed(t *testing.T) {
	client := newChartRepositoryClient()
	defer client.CloseIdleConnections()
	for _, repository := range []string{"file:///tmp/chart", "s3://example.com/charts", "@acme", "alias:acme"} {
		_, err := fetchRepositoryChart(context.Background(), client, repository, "acme", "1.0.0")
		require.Error(t, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := fetchRepositoryChart(ctx, client, chartRepositoryFixture(t, false, false), "acme", "1.0.0")
	require.ErrorIs(t, err, context.Canceled)
}

type chartResultClient struct {
	nuonrunner.Client
	results []*models.ServiceCreateRunnerJobExecutionResultRequest
	err     error
}

func (c *chartResultClient) CreateJobExecutionResult(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	c.results = append(c.results, req)
	return &models.AppRunnerJobExecutionResult{}, c.err
}

type chartStoreCopier struct {
	ocicopy.Copier
	calls atomic.Int32
}

func (c *chartStoreCopier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, tag string, _ *configs.OCIRegistryRepository, _ string) (*ocispec.Descriptor, error) {
	c.calls.Add(1)
	desc, err := store.Resolve(ctx, tag)
	if err != nil {
		return nil, err
	}
	reader, err := store.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	return &desc, reader.Close()
}

func TestHelmExecRepositoryWithMemoryAndReplacedOSRoots(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	for _, oci := range []bool{false, true} {
		for _, backend := range []string{"memory", "os"} {
			t.Run(fmt.Sprintf("oci=%t/%s", oci, backend), func(t *testing.T) {
				var fs afero.Fs = afero.NewMemMapFs()
				if backend == "os" {
					fs = afero.NewOsFs()
				}
				api, copier := &chartResultClient{}, &chartStoreCopier{}
				h, err := New(HandlerParams{V: validator.New(), FS: fs, APIClient: api, OCICopy: copier})
				require.NoError(t, err)
				h.state = &handlerState{plan: &plantypes.BuildPlan{}, cfg: &plantypes.HelmBuildPlan{HelmRepoConfig: &config.HelmRepoConfig{
					RepoURL: chartRepositoryFixture(t, oci, false), Chart: "acme", Version: "^1.0.0",
				}}}
				ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
				job, execution := &models.AppRunnerJob{}, &models.AppRunnerJobExecution{ID: fmt.Sprintf("helm-repository-%d", time.Now().UnixNano())}
				var movedRoots []string
				t.Cleanup(func() {
					require.NoError(t, h.Cleanup(ctx, job, execution))
					for _, root := range movedRoots {
						require.NoError(t, os.RemoveAll(root))
					}
				})
				require.NoError(t, h.Initialize(ctx, job, execution))
				outside := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(outside, "canary"), []byte("untouched"), 0600))
				for _, name := range []string{"HELM_REPOSITORY_CONFIG", "HELM_REPOSITORY_CACHE", "HELM_REGISTRY_CONFIG", "HELM_CONTENT_CACHE", "HELM_PLUGINS", "DOCKER_CONFIG"} {
					t.Setenv(name, outside)
				}
				if backend == "os" {
					for _, root := range []string{h.state.workspace.Root(), h.state.arch.TmpDir()} {
						moved := root + "-moved"
						require.NoError(t, os.Rename(root, moved))
						movedRoots = append(movedRoots, moved)
						require.NoError(t, os.Symlink(outside, root))
					}
				}
				require.NoError(t, h.Exec(ctx, job, execution))
				require.EqualValues(t, 1, copier.calls.Load())
				require.Len(t, api.results, 1)
				require.True(t, api.results[0].Success)
				require.NotEmpty(t, api.results[0].ContentsCompressed)
				require.Equal(t, "1.2.3+meta", h.state.chart.Metadata.Version)
				require.Len(t, h.state.chart.Dependencies(), 1)
				require.Equal(t, "0.1.2", h.state.chart.Lock.Dependencies[0].Version)
				packaged, err := h.state.output.ReadFile(defaultChartPackageFilename)
				require.NoError(t, err)
				_, err = chartArchiveFiles(bytes.NewReader(packaged))
				require.NoError(t, err)
				require.NoError(t, h.Cleanup(ctx, job, execution))
				entries, err := os.ReadDir(outside)
				require.NoError(t, err)
				require.Len(t, entries, 1, "no library config, cache, plugin, archive, or package writes may follow display paths")
				canary, err := os.ReadFile(filepath.Join(outside, "canary"))
				require.NoError(t, err)
				require.Equal(t, "untouched", string(canary))
			})
		}
	}
}

func TestHelmExecPreservesResultReportingSemantics(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	core, logs := observer.New(zap.ErrorLevel)
	l := zap.New(core)
	api := &chartResultClient{err: errors.New("result API unavailable")}
	copier := &chartStoreCopier{}
	h, err := New(HandlerParams{
		V: validator.New(), FS: afero.NewMemMapFs(), APIClient: api, OCICopy: copier,
		ErrRecorder: errs.NewRecorder(errs.Params{L: l, LC: fxtest.NewLifecycle(t)}),
	})
	require.NoError(t, err)
	h.state = &handlerState{plan: &plantypes.BuildPlan{}, cfg: &plantypes.HelmBuildPlan{HelmRepoConfig: &config.HelmRepoConfig{
		RepoURL: chartRepositoryFixture(t, false, false), Chart: "acme", Version: "1.0.0",
	}}}
	ctx := pkgctx.SetLogger(context.Background(), l)
	job, execution := &models.AppRunnerJob{ID: "job-acme"}, &models.AppRunnerJobExecution{ID: fmt.Sprintf("helm-result-%d", time.Now().UnixNano())}
	t.Cleanup(func() { require.NoError(t, h.Cleanup(ctx, job, execution)) })
	require.NoError(t, h.Initialize(ctx, job, execution))

	replacement := &chartResultClient{err: api.err}
	require.Same(t, api, h.SetAPIClient(replacement))
	require.NoError(t, h.Exec(ctx, job, execution), "result-reporting failures remain recorded, not returned")
	require.Empty(t, api.results)
	require.Len(t, replacement.results, 1)
	require.True(t, replacement.results[0].Success)
	require.NotEmpty(t, replacement.results[0].ContentsCompressed)
	require.EqualValues(t, 1, copier.calls.Load())
	require.Equal(t, 1, logs.FilterMessage("write job execution result").Len())
}

func TestRepositoryRetryKeepsClaimedChartVersion(t *testing.T) {
	upstream, err := url.Parse(strings.TrimSuffix(chartRepositoryFixture(t, false, false), "/charts"))
	require.NoError(t, err)
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	var advanced, overwritten atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/charts/index.yaml" {
			versions := []string{"1.0.0"}
			if advanced.Load() {
				versions = append(versions, "1.2.3+meta")
			}
			entries := make([]map[string]any, 0, len(versions))
			for _, version := range versions {
				entries = append(entries, map[string]any{"apiVersion": "v2", "name": "acme", "version": version, "urls": []string{"archives/acme-" + version + ".tgz"}})
			}
			json.NewEncoder(w).Encode(map[string]any{"apiVersion": "v1", "entries": map[string]any{"acme": entries}})
			return
		}
		if overwritten.Load() && r.URL.Path == "/charts/archives/acme-1.0.0.tgz" {
			w.Write(chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: "apiVersion: v2\nname: acme\nversion: 1.0.0\ndescription: overwritten\n"}))
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
	api := &chartResultClient{}
	packageChart := func(identity *sourceidentity.Identity) (*handler, error) {
		h, err := New(HandlerParams{V: validator.New(), FS: afero.NewMemMapFs(), APIClient: api})
		require.NoError(t, err)
		h.state = &handlerState{plan: &plantypes.BuildPlan{SourceIdentity: identity}, cfg: &plantypes.HelmBuildPlan{HelmRepoConfig: &config.HelmRepoConfig{
			RepoURL: server.URL + "/charts", Chart: "acme", Version: "^1.0.0",
		}}}
		job, execution := &models.AppRunnerJob{ID: "job"}, &models.AppRunnerJobExecution{ID: fmt.Sprintf("retry-%d", time.Now().UnixNano())}
		require.NoError(t, h.Initialize(ctx, job, execution))
		t.Cleanup(func() { require.NoError(t, h.Cleanup(ctx, job, execution)) })
		_, err = h.packageChart(ctx, zap.NewNop())
		return h, err
	}
	prepared := sourceidentity.New("helm", "config")
	prepared.MarkUnknown("source_requires_runner_resolution")
	prepared.Seal()
	first, err := packageChart(prepared)
	require.NoError(t, err)
	require.Equal(t, "1.0.0", first.state.chart.Metadata.Version)
	require.Len(t, api.results, 1)
	data, err := json.Marshal(api.results[0].SourceCapture)
	require.NoError(t, err)
	var resumed sourceidentity.Identity
	require.NoError(t, json.Unmarshal(data, &resumed))
	require.Equal(t, "1.0.0", resumed.ChartVersion)
	require.True(t, sourceidentity.Matches(first.state.plan.SourceIdentity, &resumed))
	advanced.Store(true)
	retried, err := packageChart(&resumed)
	require.NoError(t, err)
	require.Equal(t, "1.0.0", retried.state.chart.Metadata.Version)
	fresh, err := packageChart(prepared)
	require.NoError(t, err)
	require.Equal(t, "1.2.3+meta", fresh.state.chart.Metadata.Version)
	overwritten.Store(true)
	_, err = packageChart(&resumed)
	require.ErrorContains(t, err, "remote chart differs from captured source")
}
