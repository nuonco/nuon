package bundle

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

func TestExtractOpenVerifyAndCopy(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	logical := manifestFor(root)
	var archive bytes.Buffer
	result, err := Generate(ctx, &archive, logical, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)

	dir := t.TempDir()
	checksum, err := Extract(dir, bytes.NewReader(archive.Bytes()))
	require.NoError(t, err)
	require.Equal(t, result.TransportSHA256, checksum)
	b, err := Open(ctx, dir)
	require.NoError(t, err)
	require.Equal(t, logical, b.Manifest)
	require.Equal(t, []Member{{Key: "component:fixture", Kind: "component", Name: "fixture", MediaType: root.MediaType, Digest: root.Digest, Size: root.Size}}, b.Members())
	require.Equal(t, []Root{{Descriptor: root, Source: store}}[0].Descriptor, b.Roots[0])
	require.NoError(t, VerifyBlobs(dir))

	dst := memory.New()
	_, err = oras.Copy(ctx, b.Store(), root.Digest.String(), dst, root.Digest.String(), oras.DefaultCopyOptions)
	require.NoError(t, err)

	path := filepath.Join(dir, "blobs", "sha256", root.Digest.Encoded())
	require.NoError(t, os.WriteFile(path, bytes.Repeat([]byte("x"), int(root.Size)), 0644))
	require.Error(t, VerifyBlobs(dir))
}

func TestVerifyBlobsRejectsMissingReachableBlob(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	config := descriptor(ocispec.MediaTypeImageConfig, []byte(`{"architecture":"amd64","os":"linux"}`))
	layer := descriptor("application/vnd.oci.image.layer.v1.tar", []byte("layer data"))
	require.NoError(t, store.Push(ctx, config, bytes.NewReader([]byte(`{"architecture":"amd64","os":"linux"}`))))
	require.NoError(t, store.Push(ctx, layer, bytes.NewReader([]byte("layer data"))))
	manifestBytes := mustJSON(t, ocispec.Manifest{Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageManifest, Config: config, Layers: []ocispec.Descriptor{layer}})
	root := descriptor(ocispec.MediaTypeImageManifest, manifestBytes)
	require.NoError(t, store.Push(ctx, root, bytes.NewReader(manifestBytes)))

	var archive bytes.Buffer
	_, err := Generate(ctx, &archive, manifestFor(root), []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)

	for _, missing := range []struct {
		name string
		desc ocispec.Descriptor
	}{{name: "config", desc: config}, {name: "layer", desc: layer}} {
		t.Run(missing.name, func(t *testing.T) {
			dir := t.TempDir()
			_, err := Extract(dir, bytes.NewReader(archive.Bytes()))
			require.NoError(t, err)
			require.NoError(t, os.Remove(filepath.Join(dir, blobPath(missing.desc))))
			require.ErrorContains(t, VerifyBlobs(dir), "required blob "+missing.desc.Digest.String())
		})
	}
}

func TestVerifyBlobGraphBoundsMetadataAndRejectsConflictingSize(t *testing.T) {
	dir := t.TempDir()
	contents := bytes.Repeat([]byte("x"), 1024*1024)
	desc := descriptor(ocispec.MediaTypeImageManifest, contents)
	desc.Size = 2
	path := filepath.Join(dir, blobPath(desc))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, contents, 0644))
	require.ErrorContains(t, verifyBlobGraph(dir, desc, make(map[traversalKey]int64)), "expected 2, got 3")

	desc.Size = int64(len(contents))
	traversed := map[traversalKey]int64{{digest: desc.Digest, mediaType: desc.MediaType}: desc.Size}
	conflicting := desc
	conflicting.Size++
	require.ErrorContains(t, verifyBlobGraph(dir, conflicting, traversed), "conflicting size")
}

func TestExtractRejectsParentPath(t *testing.T) {
	_, err := Extract(t.TempDir(), archiveWithEntries(t, tarEntry{name: "../evil", contents: "x"}))
	require.ErrorContains(t, err, "unsafe path")
}

func TestExtractRejectsNonCanonicalAndDuplicatePaths(t *testing.T) {
	for _, name := range []string{"./file", "dir/../file", "dir//file", `dir\file`} {
		t.Run(name, func(t *testing.T) {
			_, err := Extract(t.TempDir(), archiveWithEntries(t, tarEntry{name: name, contents: "x"}))
			require.ErrorContains(t, err, "unsafe path")
		})
	}

	dir := t.TempDir()
	_, err := Extract(dir, archiveWithEntries(t,
		tarEntry{name: "file", contents: "first"},
		tarEntry{name: "file", contents: "second"},
	))
	require.ErrorContains(t, err, "duplicate path")
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	require.Empty(t, entries)
}

func TestExtractEnforcesLimits(t *testing.T) {
	defaults := ExtractOptions{MaxEntries: 10, MaxFileBytes: 10, MaxTotalBytes: 10, MaxDecoderMemory: DefaultMaxDecoderMemory, MaxDecoderWindow: DefaultMaxDecoderWindow}
	tests := []struct {
		name    string
		entries []tarEntry
		update  func(*ExtractOptions)
		want    string
	}{
		{name: "entries", entries: []tarEntry{{name: "a", contents: "x"}, {name: "b", contents: "x"}}, update: func(o *ExtractOptions) { o.MaxEntries = 1 }, want: "entry count"},
		{name: "file", entries: []tarEntry{{name: "a", contents: "xx"}}, update: func(o *ExtractOptions) { o.MaxFileBytes = 1 }, want: "size 2 exceeds"},
		{name: "aggregate", entries: []tarEntry{{name: "a", contents: "xx"}, {name: "b", contents: "xx"}}, update: func(o *ExtractOptions) { o.MaxTotalBytes = 3 }, want: "expanded size"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaults
			tt.update(&opts)
			dir := t.TempDir()
			_, err := ExtractWithOptions(dir, archiveWithEntries(t, tt.entries...), opts)
			require.ErrorContains(t, err, tt.want)
			entries, readErr := os.ReadDir(dir)
			require.NoError(t, readErr)
			require.Empty(t, entries)
		})
	}
}

func TestExtractRejectsUnsupportedEntryTypeAndNonEmptyDestination(t *testing.T) {
	_, err := Extract(t.TempDir(), archiveWithEntries(t, tarEntry{name: "link", entryType: tar.TypeSymlink}))
	require.ErrorContains(t, err, "non-regular entry")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "existing"), []byte("safe"), 0644))
	_, err = Extract(dir, archiveWithEntries(t, tarEntry{name: "file", contents: "x"}))
	require.ErrorContains(t, err, "must be empty")
	require.FileExists(t, filepath.Join(dir, "existing"))
}

type tarEntry struct {
	name      string
	contents  string
	entryType byte
}

func archiveWithEntries(t *testing.T, entries ...tarEntry) io.Reader {
	t.Helper()
	var compressed bytes.Buffer
	zw, err := zstd.NewWriter(&compressed)
	require.NoError(t, err)
	tw := tar.NewWriter(zw)
	for _, entry := range entries {
		entryType := entry.entryType
		if entryType == 0 {
			entryType = tar.TypeReg
		}
		size := int64(len(entry.contents))
		if entryType != tar.TypeReg && entryType != tar.TypeRegA {
			size = 0
		}
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: entry.name, Typeflag: entryType, Size: size, Mode: 0644}))
		if size > 0 {
			_, err = tw.Write([]byte(entry.contents))
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	require.NoError(t, zw.Close())
	return bytes.NewReader(compressed.Bytes())
}
