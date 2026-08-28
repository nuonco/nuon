package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/memory"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type testBundle struct {
	archive  []byte
	index    []byte
	checksum string
	blobs    map[digest.Digest][]byte
}

func buildTestBundle(t *testing.T) testBundle {
	t.Helper()
	ctx := context.Background()
	store := memory.New()
	layerData := []byte("pull bundle artifact data")
	layer := ocispec.Descriptor{MediaType: "application/octet-stream", Digest: digest.FromBytes(layerData), Size: int64(len(layerData))}
	if err := store.Push(ctx, layer, bytes.NewReader(layerData)); err != nil {
		t.Fatal(err)
	}
	manifestJSON, err := json.Marshal(ocispec.Manifest{Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageManifest, Config: layer, Layers: []ocispec.Descriptor{layer}})
	if err != nil {
		t.Fatal(err)
	}
	root := ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: digest.FromBytes(manifestJSON), Size: int64(len(manifestJSON))}
	if err := store.Push(ctx, root, bytes.NewReader(manifestJSON)); err != nil {
		t.Fatal(err)
	}
	logical := bundle.LogicalManifest{
		SchemaVersion: 1,
		Target:        bundle.Target{OS: "linux", Architecture: "amd64"},
		Components:    []bundle.Component{{Name: "fixture", Type: "fixture", ConfigDigest: digest.FromString("config").String(), Artifact: bundle.Artifact{MediaType: root.MediaType, Digest: root.Digest.String(), Size: root.Size}}},
	}
	blobs := map[digest.Digest][]byte{}
	var archive bytes.Buffer
	result, err := bundle.GenerateWithOptions(ctx, &archive, logical, bundle.Documents{PlanEnvelope: json.RawMessage(`{"schema_version":1}`)}, []bundle.Root{{Descriptor: root, Source: store}}, bundle.GenerateOptions{
		BlobSink: func(dgst digest.Digest, data []byte) error {
			blobs[dgst] = append([]byte(nil), data...)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	blobs[digest.FromBytes(result.Index)] = append([]byte(nil), result.Index...)
	return testBundle{archive: archive.Bytes(), index: result.Index, checksum: result.TransportSHA256, blobs: blobs}
}

type testBlobServer struct {
	tb       testBundle
	server   *httptest.Server
	hits     atomic.Int64
	checksum string
}

func newTestBlobServer(t *testing.T, tb testBundle) *testBlobServer {
	t.Helper()
	s := &testBlobServer{tb: tb, checksum: tb.checksum}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("presigned request included authorization header")
		}
		dgst := digest.Digest("sha256:" + filepath.Base(r.URL.Path))
		data, ok := tb.blobs[dgst]
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.hits.Add(1)
		_, _ = w.Write(data)
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *testBlobServer) grants(_ context.Context, digests []string) (*models.ServiceBlobGrantsResponse, error) {
	if len(digests) > blobGrantBatchSize {
		return nil, fmt.Errorf("too many digests requested: %d", len(digests))
	}
	resp := &models.ServiceBlobGrantsResponse{
		OciIndexDigest:    digest.FromBytes(s.tb.index).String(),
		TransportChecksum: "sha256:" + s.checksum,
	}
	for _, d := range digests {
		dgst := digest.Digest(d)
		data, ok := s.tb.blobs[dgst]
		if !ok {
			return nil, fmt.Errorf("unknown blob %s", d)
		}
		resp.Grants = append(resp.Grants, &models.ServiceBlobGrantItem{Digest: d, URL: s.server.URL + "/" + dgst.Encoded(), Size: int64(len(data))})
	}
	return resp, nil
}

func TestPullBundleFirstPull(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	dir := t.TempDir()
	destination := filepath.Join(dir, "bundle.tar.zst")

	stats, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: filepath.Join(dir, "cache")})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, tb.archive)
	if stats.CachedBlobs != 0 {
		t.Errorf("CachedBlobs = %d, want 0", stats.CachedBlobs)
	}
	if stats.DownloadedBlobs != stats.TotalBlobs {
		t.Errorf("DownloadedBlobs = %d, want %d", stats.DownloadedBlobs, stats.TotalBlobs)
	}
	if stats.TotalBlobs != len(tb.blobs)-1 {
		t.Errorf("TotalBlobs = %d, want %d (all blobs except the index)", stats.TotalBlobs, len(tb.blobs)-1)
	}
}

func TestPullBundleSecondPullUsesCache(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")

	if _, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: filepath.Join(dir, "first.tar.zst"), CacheDir: cacheDir}); err != nil {
		t.Fatal(err)
	}
	srv.hits.Store(0)

	destination := filepath.Join(dir, "second.tar.zst")
	stats, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: cacheDir})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, tb.archive)
	if stats.DownloadedBlobs != 0 {
		t.Errorf("DownloadedBlobs = %d, want 0", stats.DownloadedBlobs)
	}
	if stats.CachedBlobs != stats.TotalBlobs {
		t.Errorf("CachedBlobs = %d, want %d", stats.CachedBlobs, stats.TotalBlobs)
	}
	if hits := srv.hits.Load(); hits != 0 {
		t.Errorf("blob server hits = %d, want 0", hits)
	}
}

func TestPullBundlePartialCache(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")

	cache := blobCache{dir: cacheDir}
	var seeded int
	for dgst, data := range tb.blobs {
		if dgst == digest.FromBytes(tb.index) {
			continue
		}
		if err := cache.put(dgst, data); err != nil {
			t.Fatal(err)
		}
		seeded++
		if seeded == 2 {
			break
		}
	}

	destination := filepath.Join(dir, "bundle.tar.zst")
	stats, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: cacheDir})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, tb.archive)
	if stats.CachedBlobs != seeded {
		t.Errorf("CachedBlobs = %d, want %d", stats.CachedBlobs, seeded)
	}
	if stats.DownloadedBlobs != stats.TotalBlobs-seeded {
		t.Errorf("DownloadedBlobs = %d, want %d", stats.DownloadedBlobs, stats.TotalBlobs-seeded)
	}
}

func TestPullBundleHealsCorruptCachedBlob(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")

	if _, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: filepath.Join(dir, "first.tar.zst"), CacheDir: cacheDir}); err != nil {
		t.Fatal(err)
	}

	corrupted := digest.FromBytes([]byte("pull bundle artifact data"))
	if _, ok := tb.blobs[corrupted]; !ok {
		t.Fatal("leaf blob to corrupt not found in bundle")
	}
	cache := blobCache{dir: cacheDir}
	corrupt := bytes.Repeat([]byte("x"), len(tb.blobs[corrupted]))
	if err := os.WriteFile(cache.path(corrupted), corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	destination := filepath.Join(dir, "second.tar.zst")
	if _, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: cacheDir}); err != nil {
		t.Fatal(err)
	}
	assertFileContent(t, destination, tb.archive)
	restored, err := cache.read(corrupted)
	if err != nil {
		t.Fatal(err)
	}
	if digest.FromBytes(restored) != corrupted {
		t.Error("corrupt cached blob was not healed")
	}
}

func TestPullBundleChecksumMismatch(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	srv.checksum = "0000000000000000000000000000000000000000000000000000000000000000"
	dir := t.TempDir()
	destination := filepath.Join(dir, "bundle.tar.zst")

	_, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: filepath.Join(dir, "cache")})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err = %v, want checksum mismatch", err)
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Error("destination file exists after checksum mismatch")
	}
	if _, statErr := os.Stat(destination + ".partial"); !os.IsNotExist(statErr) {
		t.Error("partial file exists after checksum mismatch")
	}
}

func TestPullBundleRefusesExistingDestination(t *testing.T) {
	tb := buildTestBundle(t)
	srv := newTestBlobServer(t, tb)
	dir := t.TempDir()
	destination := filepath.Join(dir, "bundle.tar.zst")
	if err := os.WriteFile(destination, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := pullBundle(context.Background(), srv.server.Client(), srv.grants, PullBundleOptions{File: destination, CacheDir: filepath.Join(dir, "cache")})
	if err == nil || !strings.Contains(err.Error(), "use --overwrite") {
		t.Fatalf("err = %v, want overwrite refusal", err)
	}
	assertFileContent(t, destination, []byte("existing"))
}
