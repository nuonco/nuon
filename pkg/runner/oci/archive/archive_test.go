package ociarchive

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
)

func TestPackUnpackRoundTrip(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	files := []FileRef{
		writeTestFile(t, sourceDir, "nested/main.tf", "terraform {}", 0644),
		writeTestFile(t, sourceDir, "bin/terraform", "executable", 0755),
		writeTestFile(t, sourceDir, "empty.txt", "", 0644),
	}

	packed := newTestArchive(t, ctx)
	if err := packed.Pack(ctx, zap.NewNop(), files); err != nil {
		t.Fatalf("Pack() error = %v", err)
	}
	manifest := resolveTestManifest(t, ctx, packed)
	layers := manifestLayers(t, ctx, packed.store, manifest)
	if len(layers) != 1 {
		t.Fatalf("layer count = %d, want 1", len(layers))
	}
	if layers[0].MediaType != tarLayerMediaType {
		t.Fatalf("layer media type = %q, want %q", layers[0].MediaType, tarLayerMediaType)
	}
	if err := validateArchiveManifest(ctx, packed.store, manifest); err != nil {
		t.Fatalf("validateArchiveManifest() error = %v", err)
	}

	unpacked := newTestArchive(t, ctx)
	_, err := oras.Copy(ctx, packed.store, defaultLocalTag, unpacked.store, defaultLocalTag, oras.DefaultCopyOptions)
	if err != nil {
		t.Fatalf("oras.Copy() error = %v", err)
	}
	for _, file := range files {
		got, err := os.ReadFile(filepath.Join(unpacked.BasePath(), filepath.FromSlash(file.RelPath)))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", file.RelPath, err)
		}
		want, err := os.ReadFile(file.AbsPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("content for %q = %q, want %q", file.RelPath, got, want)
		}
	}
	mode, err := os.Stat(filepath.Join(unpacked.BasePath(), "bin/terraform"))
	if err != nil {
		t.Fatal(err)
	}
	if mode.Mode().Perm() != 0755 {
		t.Errorf("executable mode = %o, want 755", mode.Mode().Perm())
	}
}

func TestPackDeterministic(t *testing.T) {
	ctx := context.Background()
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	firstFiles := []FileRef{
		writeTestFile(t, firstDir, "z.txt", "last", 0600),
		writeTestFile(t, firstDir, "a.txt", "first", 0750),
	}
	secondFiles := []FileRef{
		writeTestFile(t, secondDir, "a.txt", "first", 0700),
		writeTestFile(t, secondDir, "z.txt", "last", 0644),
	}
	for _, file := range secondFiles {
		if err := os.Chtimes(file.AbsPath, time.Now(), time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	first := newTestArchive(t, ctx)
	if err := first.Pack(ctx, zap.NewNop(), firstFiles); err != nil {
		t.Fatal(err)
	}
	second := newTestArchive(t, ctx)
	if err := second.Pack(ctx, zap.NewNop(), secondFiles); err != nil {
		t.Fatal(err)
	}

	firstDigest := resolveTestManifest(t, ctx, first).Digest
	secondDigest := resolveTestManifest(t, ctx, second).Digest
	if firstDigest != secondDigest {
		t.Fatalf("manifest digests differ: %s != %s", firstDigest, secondDigest)
	}
}

func TestValidateArchiveManifestRejectsLegacyPerFileLayers(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	legacyStore, err := file.New(filepath.Join(t.TempDir(), "legacy-store"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = legacyStore.Close() })

	legacyFile := writeTestFile(t, sourceDir, "modules/example/main.tf", "terraform {}", 0644)
	layer, err := legacyStore.Add(ctx, legacyFile.RelPath, "application/octet-stream", legacyFile.AbsPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := oras.Pack(ctx, legacyStore, defaultArtifactType, []v1.Descriptor{layer}, oras.PackOptions{
		PackImageManifest: true,
		ManifestAnnotations: map[string]string{
			v1.AnnotationCreated: "1970-01-01T00:00:00Z",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyStore.Tag(ctx, manifest, defaultLocalTag); err != nil {
		t.Fatal(err)
	}

	err = validateArchiveManifest(ctx, legacyStore, manifest)
	if err == nil {
		t.Fatal("validateArchiveManifest() error = nil")
	}
	if !strings.Contains(err.Error(), "unsupported OCI archive layer format") {
		t.Fatalf("validateArchiveManifest() error = %q", err)
	}
}

func TestPackManifestRejectsExcessLayers(t *testing.T) {
	files := make([]FileRef, 1285)
	layers := make([]v1.Descriptor, maxManifestLayers+1)
	_, err := packManifest(context.Background(), nil, files, layers)
	if err == nil {
		t.Fatal("packManifest() error = nil")
	}
	if !strings.Contains(err.Error(), "1285 files into 801 OCI layers") {
		t.Fatalf("packManifest() error = %q", err)
	}
}

func newTestArchive(t *testing.T, ctx context.Context) *archive {
	t.Helper()
	a := New()
	if err := a.Initialize(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Cleanup(ctx) })
	return a
}

func writeTestFile(t *testing.T, root, relPath, contents string, mode os.FileMode) FileRef {
	t.Helper()
	absPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absPath, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
	return FileRef{AbsPath: absPath, RelPath: relPath, FileType: "application/octet-stream"}
}

func resolveTestManifest(t *testing.T, ctx context.Context, a *archive) v1.Descriptor {
	t.Helper()
	manifest, err := a.store.Resolve(ctx, defaultLocalTag)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func manifestLayers(t *testing.T, ctx context.Context, store *file.Store, manifest v1.Descriptor) []v1.Descriptor {
	t.Helper()
	data, err := content.FetchAll(ctx, store, manifest)
	if err != nil {
		t.Fatal(err)
	}
	var parsed v1.Manifest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	return parsed.Layers
}
