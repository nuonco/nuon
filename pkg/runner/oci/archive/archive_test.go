package ociarchive

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
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
	if len(layers) != 2 {
		t.Fatalf("layer count = %d, want 2", len(layers))
	}
	for _, layer := range layers {
		if layer.MediaType != tarLayerMediaType {
			t.Fatalf("layer media type = %q, want %q", layer.MediaType, tarLayerMediaType)
		}
	}

	unpacked := newTestArchive(t, ctx)
	_, err := oras.Copy(ctx, packed.store, defaultLocalTag, unpacked.store, defaultLocalTag, oras.DefaultCopyOptions)
	if err != nil {
		t.Fatalf("oras.Copy() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(unpacked.BasePath(), "nested/main.tf")); err != nil {
		t.Fatalf("legacy reader did not auto-unpack new tar layers: %v", err)
	}
	for _, file := range files[:2] {
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
	if _, err := os.Stat(filepath.Join(unpacked.BasePath(), "empty.txt")); !os.IsNotExist(err) {
		t.Fatalf("empty file stat error = %v, want not found", err)
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

func TestPackReusesUnchangedDirectoryLayer(t *testing.T) {
	ctx := context.Background()
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	first := newTestArchive(t, ctx)
	if err := first.Pack(ctx, zap.NewNop(), []FileRef{
		writeTestFile(t, firstDir, "stable/main.tf", "unchanged", 0644),
		writeTestFile(t, firstDir, "changed/main.tf", "before", 0644),
	}); err != nil {
		t.Fatal(err)
	}
	second := newTestArchive(t, ctx)
	if err := second.Pack(ctx, zap.NewNop(), []FileRef{
		writeTestFile(t, secondDir, "stable/main.tf", "unchanged", 0644),
		writeTestFile(t, secondDir, "changed/main.tf", "after", 0644),
	}); err != nil {
		t.Fatal(err)
	}

	firstLayers := manifestLayersByTitle(t, ctx, first)
	secondLayers := manifestLayersByTitle(t, ctx, second)
	if firstLayers["stable"].Digest != secondLayers["stable"].Digest {
		t.Fatal("unchanged directory layer digest changed")
	}
	if firstLayers["changed"].Digest == secondLayers["changed"].Digest {
		t.Fatal("changed directory layer digest did not change")
	}
}

func TestPackUnpackDirectorySymlink(t *testing.T) {
	ctx := context.Background()
	sourceDir := t.TempDir()
	target := writeTestFile(t, sourceDir, "target/main.tf", "terraform {}", 0644)
	linkPath := filepath.Join(sourceDir, "linked")
	if err := os.Symlink("target", linkPath); err != nil {
		t.Fatal(err)
	}

	packed := newTestArchive(t, ctx)
	if err := packed.Pack(ctx, zap.NewNop(), []FileRef{
		target,
		{AbsPath: linkPath, RelPath: "linked", FileType: "application/octet-stream"},
	}); err != nil {
		t.Fatal(err)
	}
	unpacked := newTestArchive(t, ctx)
	if _, err := oras.Copy(ctx, packed.store, defaultLocalTag, unpacked.store, defaultLocalTag, oras.DefaultCopyOptions); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(filepath.Join(unpacked.BasePath(), "linked"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("linked mode = %s, want symlink", info.Mode())
	}
	contents, err := os.ReadFile(filepath.Join(unpacked.BasePath(), "linked", "main.tf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "terraform {}" {
		t.Fatalf("linked file content = %q", contents)
	}
}

func TestGroupArchiveLayersFallsBackAboveLimit(t *testing.T) {
	files := make([]FileRef, maxArchiveLayers+1)
	for i := range files {
		files[i].RelPath = filepath.ToSlash(filepath.Join("dir-"+strconv.Itoa(i), "file"))
	}

	atLimit := groupArchiveLayers(files[:maxArchiveLayers])
	if len(atLimit) != maxArchiveLayers {
		t.Fatalf("layer count at limit = %d, want %d", len(atLimit), maxArchiveLayers)
	}

	layers := groupArchiveLayers(files)
	if len(layers) != 1 {
		t.Fatalf("layer count = %d, want 1", len(layers))
	}
	if layers[0].title != tarLayerTitle {
		t.Fatalf("layer title = %q, want %q", layers[0].title, tarLayerTitle)
	}
	if len(layers[0].files) != len(files) {
		t.Fatalf("layer file count = %d, want %d", len(layers[0].files), len(files))
	}
}

func TestUnpackLegacyPerFileLayers(t *testing.T) {
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

	unpacked := newTestArchive(t, ctx)
	_, err = oras.Copy(ctx, legacyStore, defaultLocalTag, unpacked.store, defaultLocalTag, oras.DefaultCopyOptions)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(unpacked.BasePath(), filepath.FromSlash(legacyFile.RelPath)))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "terraform {}" {
		t.Fatalf("legacy file content = %q", got)
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

func manifestLayersByTitle(t *testing.T, ctx context.Context, archive *archive) map[string]v1.Descriptor {
	t.Helper()
	layers := manifestLayers(t, ctx, archive.store, resolveTestManifest(t, ctx, archive))
	byTitle := make(map[string]v1.Descriptor, len(layers))
	for _, layer := range layers {
		byTitle[layer.Annotations[v1.AnnotationTitle]] = layer
	}
	return byTitle
}
