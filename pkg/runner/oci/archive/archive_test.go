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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"

	"github.com/nuonco/nuon/pkg/sourcefs"
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
	if err := packed.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), sourceDir), files); err != nil {
		t.Fatalf("Pack() error = %v", err)
	}
	manifest := resolveTestManifest(t, ctx, packed)
	layers := manifestLayers(t, ctx, packed.Ref(), manifest)
	if len(layers) != 3 {
		t.Fatalf("layer count = %d, want 3", len(layers))
	}
	for _, layer := range layers {
		if layer.MediaType != tarLayerMediaType {
			t.Fatalf("layer media type = %q, want %q", layer.MediaType, tarLayerMediaType)
		}
	}

	unpacked := newTestArchive(t, ctx)
	_, err := oras.Copy(ctx, packed.Ref(), defaultLocalTag, unpacked.Ref().(oras.Target), defaultLocalTag, oras.DefaultCopyOptions)
	if err != nil {
		t.Fatalf("oras.Copy() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(unpacked.BasePath(), "nested/main.tf")); err != nil {
		t.Fatalf("legacy reader did not auto-unpack new tar layers: %v", err)
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
	if err := first.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), firstDir), firstFiles); err != nil {
		t.Fatal(err)
	}
	second := newTestArchive(t, ctx)
	if err := second.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), secondDir), secondFiles); err != nil {
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
	if err := first.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), firstDir), []FileRef{
		writeTestFile(t, firstDir, "stable/main.tf", "unchanged", 0644),
		writeTestFile(t, firstDir, "changed/main.tf", "before", 0644),
	}); err != nil {
		t.Fatal(err)
	}
	second := newTestArchive(t, ctx)
	if err := second.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), secondDir), []FileRef{
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
	if err := packed.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), sourceDir), []FileRef{
		target,
		{AbsPath: linkPath, RelPath: "linked", FileType: "application/octet-stream"},
	}); err != nil {
		t.Fatal(err)
	}
	unpacked := newTestArchive(t, ctx)
	if _, err := oras.Copy(ctx, packed.Ref(), defaultLocalTag, unpacked.Ref().(oras.Target), defaultLocalTag, oras.DefaultCopyOptions); err != nil {
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

func TestPackLayerLimit(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	files := make([]FileRef, maxArchiveLayers+1)
	for i := range files {
		files[i] = writeInputFile(t, fs, "/source", "dir-"+strconv.Itoa(i)+"/file", "content", 0644)
	}
	source := newTestSource(t, fs, "/source")
	atLimit := newTestArchive(t, ctx)
	require.NoError(t, atLimit.Pack(ctx, zap.NewNop(), source, files[:maxArchiveLayers]))
	require.Len(t, manifestLayers(t, ctx, atLimit.Ref(), resolveTestManifest(t, ctx, atLimit)), maxArchiveLayers)
	aboveLimit := newTestArchive(t, ctx)
	require.NoError(t, aboveLimit.Pack(ctx, zap.NewNop(), source, files))
	layers := manifestLayers(t, ctx, aboveLimit.Ref(), resolveTestManifest(t, ctx, aboveLimit))
	require.Len(t, layers, 1)
	require.Equal(t, tarLayerTitle, layers[0].Annotations[v1.AnnotationTitle])
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
	_, err = oras.Copy(ctx, legacyStore, defaultLocalTag, unpacked.Ref().(oras.Target), defaultLocalTag, oras.DefaultCopyOptions)
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
	t.Cleanup(func() { require.NoError(t, a.Cleanup(ctx)) })
	return a
}

func writeTestFile(t *testing.T, root, relPath, contents string, mode os.FileMode) FileRef {
	t.Helper()
	return writeInputFile(t, afero.NewOsFs(), root, relPath, contents, mode)
}

func writeInputFile(t *testing.T, fs afero.Fs, root, relPath, contents string, mode os.FileMode) FileRef {
	t.Helper()
	absPath := filepath.Join(root, filepath.FromSlash(relPath))
	require.NoError(t, fs.MkdirAll(filepath.Dir(absPath), 0755))
	require.NoError(t, afero.WriteFile(fs, absPath, []byte(contents), mode))
	require.NoError(t, fs.Chmod(absPath, mode))
	return FileRef{AbsPath: absPath, RelPath: relPath, FileType: "application/octet-stream"}
}

func newTestSource(t *testing.T, fs afero.Fs, root string) *sourcefs.FS {
	t.Helper()
	source, err := sourcefs.New(fs, root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	return source
}

func resolveTestManifest(t *testing.T, ctx context.Context, a *archive) v1.Descriptor {
	t.Helper()
	manifest, err := a.Ref().Resolve(ctx, defaultLocalTag)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func manifestLayers(t *testing.T, ctx context.Context, store content.Fetcher, manifest v1.Descriptor) []v1.Descriptor {
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
	layers := manifestLayers(t, ctx, archive.Ref(), resolveTestManifest(t, ctx, archive))
	byTitle := make(map[string]v1.Descriptor, len(layers))
	for _, layer := range layers {
		byTitle[layer.Annotations[v1.AnnotationTitle]] = layer
	}
	return byTitle
}
