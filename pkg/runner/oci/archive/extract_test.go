package ociarchive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/errdef"
)

type tarEntry struct {
	header tar.Header
	data   string
}

type fixtureLayer struct {
	title    string
	data     []byte
	unpack   bool
	checksum string
}

func localFixture(t *testing.T, layers ...fixtureLayer) *memory.Store {
	t.Helper()
	ctx := context.Background()
	store := memory.New()
	descriptors := make([]v1.Descriptor, 0, len(layers))
	for _, layer := range layers {
		desc := content.NewDescriptorFromBytes("application/octet-stream", layer.data)
		desc.Annotations = map[string]string{v1.AnnotationTitle: layer.title}
		if layer.unpack {
			// Old file.Store directories use this standard media type, not Nuon's.
			desc.MediaType = v1.MediaTypeImageLayerGzip
			desc.Annotations[file.AnnotationUnpack] = "true"
		}
		if layer.checksum != "" {
			desc.Annotations[file.AnnotationDigest] = layer.checksum
		}
		err := store.Push(ctx, desc, bytes.NewReader(layer.data))
		if !errors.Is(err, errdef.ErrAlreadyExists) {
			require.NoError(t, err)
		}
		descriptors = append(descriptors, desc)
	}
	manifest, err := oras.Pack(ctx, store, defaultArtifactType, descriptors, oras.PackOptions{PackImageManifest: true})
	require.NoError(t, err)
	require.NoError(t, store.Tag(ctx, manifest, defaultLocalTag))
	return store
}

func fixtureTar(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		h := entry.header
		if h.Typeflag == tar.TypeReg || h.Typeflag == tar.TypeRegA {
			h.Size = int64(len(entry.data))
		}
		require.NoError(t, tw.WriteHeader(&h))
		_, err := tw.Write([]byte(entry.data))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return compressed.Bytes()
}

func regularEntry(name, data string) tarEntry {
	return tarEntry{header: tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0644}, data: data}
}

func linkEntry(name, target string, kind byte) tarEntry {
	return tarEntry{header: tar.Header{Name: name, Linkname: target, Typeflag: kind, Mode: 0777}}
}

func copyFixture(ctx context.Context, source oras.ReadOnlyTarget, dst *archive) error {
	_, err := oras.Copy(ctx, source, defaultLocalTag, dst.Ref().(oras.Target), defaultLocalTag, oras.DefaultCopyOptions)
	return err
}

func TestExtractionRejectsHostileNamesAndEntries(t *testing.T) {
	for _, fixture := range []string{
		"title-parent", "title-absolute", "title-dot", "title-backslash", "title-drive",
		"tar-parent", "tar-absolute", "tar-noncanonical", "tar-backslash", "tar-drive", "tar-outside-title",
		"symlink-parent", "symlink-absolute", "symlink-dot-parent", "symlink-dangling-escape", "symlink-backslash", "symlink-drive", "symlink-dot-drive",
		"symlink-chain-escape", "symlink-forward-chain-escape", "symlink-loop", "symlink-dot-loop", "symlink-self", "symlink-chain-parent",
		"symlink-parent-write", "symlink-file-write", "symlink-chain-write", "symlink-chain-directory",
		"duplicate-file", "duplicate-symlink", "file-then-symlink", "conflicting-directory", "hardlink-inside", "hardlink-outside", "hardlink-chain", "fifo", "device",
	} {
		t.Run(fixture, func(t *testing.T) {
			ctx := context.Background()
			dst := newTestArchive(t, ctx)
			canary := filepath.Join(dst.TmpDir(), "canary")
			require.NoError(t, os.WriteFile(canary, []byte("neighbor must not change"), 0640))
			layer := fixtureLayer{title: ".", unpack: true}
			var entries []tarEntry
			switch fixture {
			case "title-parent":
				layer = fixtureLayer{title: "../canary", data: []byte("bad")}
			case "title-absolute":
				layer = fixtureLayer{title: canary, data: []byte("bad")}
			case "title-dot":
				layer = fixtureLayer{title: ".", data: []byte("bad")}
			case "title-backslash":
				layer = fixtureLayer{title: `..\canary`, data: []byte("bad")}
			case "title-drive":
				layer = fixtureLayer{title: "C:/canary", data: []byte("bad")}
			case "tar-parent":
				entries = []tarEntry{regularEntry("../canary", "bad")}
			case "tar-absolute":
				entries = []tarEntry{regularEntry(canary, "bad")}
			case "tar-noncanonical":
				entries = []tarEntry{regularEntry("dir/../../canary", "bad")}
			case "tar-backslash":
				entries = []tarEntry{regularEntry(`..\canary`, "bad")}
			case "tar-drive":
				entries = []tarEntry{regularEntry("C:/canary", "bad")}
			case "tar-outside-title":
				layer.title = "nested"
				entries = []tarEntry{regularEntry("other/file", "bad")}
			case "symlink-parent":
				entries = []tarEntry{linkEntry("link", "../canary", tar.TypeSymlink)}
			case "symlink-absolute":
				entries = []tarEntry{linkEntry("link", canary, tar.TypeSymlink)}
			case "symlink-dot-parent":
				entries = []tarEntry{linkEntry("link", "./../canary", tar.TypeSymlink)}
			case "symlink-dangling-escape":
				entries = []tarEntry{linkEntry("link", "missing/../../canary", tar.TypeSymlink)}
			case "symlink-backslash":
				entries = []tarEntry{linkEntry("link", `..\canary`, tar.TypeSymlink)}
			case "symlink-drive":
				entries = []tarEntry{linkEntry("link", "C:/canary", tar.TypeSymlink)}
			case "symlink-dot-drive":
				entries = []tarEntry{linkEntry("link", "./C:/canary", tar.TypeSymlink)}
			case "symlink-chain-escape":
				entries = []tarEntry{linkEntry("first", ".", tar.TypeSymlink), linkEntry("second", "first/..", tar.TypeSymlink)}
			case "symlink-forward-chain-escape":
				entries = []tarEntry{linkEntry("first", "second/..", tar.TypeSymlink), linkEntry("second", ".", tar.TypeSymlink)}
			case "symlink-loop":
				entries = []tarEntry{linkEntry("first", "second", tar.TypeSymlink), linkEntry("second", "first", tar.TypeSymlink)}
			case "symlink-dot-loop":
				entries = []tarEntry{linkEntry("first", "./second", tar.TypeSymlink), linkEntry("second", "./first", tar.TypeSymlink)}
			case "symlink-self":
				entries = []tarEntry{linkEntry("self", "self", tar.TypeSymlink)}
			case "symlink-chain-parent":
				entries = []tarEntry{linkEntry("a", ".", tar.TypeSymlink), linkEntry("b", "a/..", tar.TypeSymlink), regularEntry("b/canary", "bad")}
			case "symlink-parent-write":
				entries = []tarEntry{linkEntry("link", "target", tar.TypeSymlink), regularEntry("link/canary", "bad")}
			case "symlink-file-write":
				entries = []tarEntry{linkEntry("link", "target", tar.TypeSymlink), regularEntry("link", "bad")}
			case "symlink-chain-write", "symlink-chain-directory":
				entries = []tarEntry{regularEntry("real/canary", "safe"), linkEntry("first", "./real", tar.TypeSymlink), linkEntry("second", "first", tar.TypeSymlink)}
				if fixture == "symlink-chain-write" {
					entries = append(entries, regularEntry("second/canary", "bad"))
				} else {
					entries = append(entries, tarEntry{header: tar.Header{Name: "second/child", Typeflag: tar.TypeDir, Mode: 0755}})
				}
			case "duplicate-file":
				entries = []tarEntry{regularEntry("file", "first"), regularEntry("file", "bad")}
			case "duplicate-symlink":
				entries = []tarEntry{linkEntry("link", "one", tar.TypeSymlink), linkEntry("link", "two", tar.TypeSymlink)}
			case "file-then-symlink":
				entries = []tarEntry{regularEntry("file", "first"), linkEntry("file", "other", tar.TypeSymlink)}
			case "conflicting-directory":
				entries = []tarEntry{{header: tar.Header{Name: "dir", Typeflag: tar.TypeDir, Mode: 0755}}, {header: tar.Header{Name: "dir", Typeflag: tar.TypeDir, Mode: 0700}}}
			case "hardlink-inside":
				entries = []tarEntry{regularEntry("first", "safe"), linkEntry("second", "first", tar.TypeLink)}
			case "hardlink-outside":
				entries = []tarEntry{linkEntry("link", canary, tar.TypeLink)}
			case "hardlink-chain":
				entries = []tarEntry{linkEntry("first", "second", tar.TypeLink), linkEntry("second", "../canary", tar.TypeLink), regularEntry("first", "bad")}
			case "fifo":
				entries = []tarEntry{{header: tar.Header{Name: "pipe", Typeflag: tar.TypeFifo, Mode: 0644}}}
			case "device":
				entries = []tarEntry{{header: tar.Header{Name: "device", Typeflag: tar.TypeChar, Mode: 0644}}}
			}
			if layer.unpack {
				layer.data = fixtureTar(t, entries...)
			}
			source := localFixture(t, layer)
			require.Error(t, copyFixture(ctx, source, dst))
			require.Error(t, copyFixture(ctx, source, dst), "retries must not publish a rejected manifest")
			if fixture == "symlink-chain-write" || fixture == "symlink-chain-directory" {
				data, err := os.ReadFile(filepath.Join(dst.BasePath(), "real/canary"))
				require.NoError(t, err)
				require.Equal(t, "safe", string(data))
				_, err = os.Stat(filepath.Join(dst.BasePath(), "real/child"))
				require.True(t, os.IsNotExist(err), "no directory may be created through a symlink chain")
			}
			_, err := dst.FS().Stat(".")
			require.Error(t, err, "rejected extraction must close the output root")
			_, err = dst.Ref().Resolve(ctx, defaultLocalTag)
			require.Error(t, err)
			contents, err := os.ReadFile(canary)
			require.NoError(t, err)
			require.Equal(t, "neighbor must not change", string(contents))
			info, err := os.Stat(canary)
			require.NoError(t, err)
			require.EqualValues(t, 0640, info.Mode().Perm())
		})
	}
}

func TestExtractionRejectsPreexistingLinksAndFiles(t *testing.T) {
	for _, fixture := range []string{"terminal-symlink", "parent-symlink", "hardlink", "existing-file"} {
		for _, unpack := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tar=%t", fixture, unpack), func(t *testing.T) {
				ctx := context.Background()
				dst := newTestArchive(t, ctx)
				outside := t.TempDir()
				canary := filepath.Join(outside, "canary")
				require.NoError(t, os.WriteFile(canary, []byte("untouched"), 0600))
				name := "canary"
				switch fixture {
				case "terminal-symlink":
					require.NoError(t, os.Symlink(canary, filepath.Join(dst.BasePath(), name)))
				case "parent-symlink":
					require.NoError(t, os.Symlink(outside, filepath.Join(dst.BasePath(), "parent")))
					name = "parent/canary"
				case "hardlink":
					require.NoError(t, os.Link(canary, filepath.Join(dst.BasePath(), name)))
				case "existing-file":
					require.NoError(t, os.WriteFile(filepath.Join(dst.BasePath(), name), []byte("original"), 0600))
				}
				layer := fixtureLayer{title: name, data: []byte("bad")}
				if unpack {
					layer = fixtureLayer{title: ".", unpack: true, data: fixtureTar(t, regularEntry(name, "bad"))}
				}
				require.Error(t, copyFixture(ctx, localFixture(t, layer), dst))
				contents, err := os.ReadFile(canary)
				require.NoError(t, err)
				require.Equal(t, "untouched", string(contents))
				if fixture == "existing-file" {
					contents, err = os.ReadFile(filepath.Join(dst.BasePath(), name))
					require.NoError(t, err)
					require.Equal(t, "original", string(contents))
				}
			})
		}
	}
}

func TestExtractionLegacyDirectoryModesOrderingAndDeduplication(t *testing.T) {
	ctx := context.Background()
	dst := newTestArchive(t, ctx)
	entries := []tarEntry{
		{header: tar.Header{Name: "nested/", Typeflag: tar.TypeDir, Mode: 0750}},
		{header: tar.Header{Name: "nested/bin/", Typeflag: tar.TypeDir, Mode: 0550}},
		{header: tar.Header{Name: "nested/bin/tool", Typeflag: tar.TypeReg, Mode: 0751}, data: "executable"},
		linkEntry("nested/link", "bin", tar.TypeSymlink),
		{header: tar.Header{Name: "nested/secret", Typeflag: tar.TypeReg, Mode: 0600}, data: "private"},
	}
	source := localFixture(t,
		fixtureLayer{title: "nested", unpack: true, data: fixtureTar(t, entries...)},
		fixtureLayer{title: "a", data: []byte("identical")},
		fixtureLayer{title: "b", data: []byte("identical")},
	)
	require.NoError(t, copyFixture(ctx, source, dst))
	require.NoError(t, copyFixture(ctx, source, dst), "repeat copy/tag is idempotent")
	for name, mode := range map[string]os.FileMode{"nested": 0750, "nested/bin": 0550, "nested/bin/tool": 0751, "nested/secret": 0600} {
		info, err := dst.FS().Stat(name)
		require.NoError(t, err)
		require.Equal(t, mode, info.Mode().Perm(), name)
	}
	for name, expected := range map[string]string{"a": "identical", "b": "identical", "nested/link/tool": "executable", "nested/secret": "private"} {
		data, err := afero.ReadFile(dst.FS(), name)
		require.NoError(t, err)
		require.Equal(t, expected, string(data))
	}
	require.NoError(t, dst.Cleanup(ctx))
}

func TestExtractionValidatesCompressedAndUncompressedIntegrity(t *testing.T) {
	valid := fixtureTar(t, regularEntry("file", "valid"))
	gz, err := gzip.NewReader(bytes.NewReader(valid))
	require.NoError(t, err)
	uncompressed, err := io.ReadAll(gz)
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	for _, fixture := range []string{"valid-digest", "wrong-digest", "invalid-digest", "bad-gzip-trailer", "truncated-gzip"} {
		t.Run(fixture, func(t *testing.T) {
			ctx := context.Background()
			dst := newTestArchive(t, ctx)
			layer := fixtureLayer{title: ".", unpack: true, data: bytes.Clone(valid)}
			switch fixture {
			case "valid-digest":
				layer.checksum = digest.FromBytes(uncompressed).String()
			case "wrong-digest":
				layer.checksum = digest.FromString("wrong").String()
			case "invalid-digest":
				layer.checksum = "not-a-digest"
			case "bad-gzip-trailer":
				layer.data[len(layer.data)-8] ^= 0xff
			case "truncated-gzip":
				layer.data = layer.data[:len(layer.data)-4]
			}
			err := copyFixture(ctx, localFixture(t, layer), dst)
			if fixture == "valid-digest" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestExtractionUsesPinnedRootAfterPathReplacement(t *testing.T) {
	ctx := context.Background()
	dst := newTestArchive(t, ctx)
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "canary"), []byte("neighbor"), 0600))
	pinned := filepath.Join(dst.TmpDir(), "pinned-output")
	require.NoError(t, os.Rename(dst.BasePath(), pinned))
	require.NoError(t, os.Symlink(outside, dst.BasePath()))
	source := localFixture(t, fixtureLayer{title: ".", unpack: true, data: fixtureTar(t, regularEntry("canary", "archive"))})
	require.NoError(t, copyFixture(ctx, source, dst))
	contents, err := afero.ReadFile(dst.FS(), "canary")
	require.NoError(t, err)
	require.Equal(t, "archive", string(contents))
	contents, err = os.ReadFile(filepath.Join(outside, "canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(contents))
	require.NoError(t, dst.Cleanup(ctx))
	_, err = dst.FS().Open("canary")
	require.Error(t, err, "Cleanup must close the borrowed root")
	contents, err = os.ReadFile(filepath.Join(outside, "canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(contents))
}

func TestPackOutputRejectsPreexistingEscapingLinks(t *testing.T) {
	ctx := context.Background()
	for _, parent := range []bool{false, true} {
		t.Run(fmt.Sprintf("parent=%t", parent), func(t *testing.T) {
			packed := newTestArchive(t, ctx)
			outside := t.TempDir()
			canary := filepath.Join(outside, "archive-v1-000.tar.gz")
			require.NoError(t, os.WriteFile(canary, []byte("neighbor"), 0600))
			if parent {
				require.NoError(t, os.Remove(filepath.Join(packed.TmpDir(), "layers")))
				require.NoError(t, os.Symlink(outside, filepath.Join(packed.TmpDir(), "layers")))
			} else {
				require.NoError(t, os.Symlink(canary, filepath.Join(packed.TmpDir(), "layers/archive-v1-000.tar.gz")))
			}
			root := t.TempDir()
			input := writeTestFile(t, root, "file", "input", 0644)
			require.Error(t, packed.Pack(ctx, zap.NewNop(), newTestSource(t, afero.NewOsFs(), root), []FileRef{input}))
			contents, err := os.ReadFile(canary)
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(contents))
		})
	}
}

func TestCleanupHandlesRestrictiveDirectoryModes(t *testing.T) {
	for _, mode := range []int64{0000, 0500, 0555} {
		t.Run(fmt.Sprintf("%o", mode), func(t *testing.T) {
			ctx := context.Background()
			dst := newTestArchive(t, ctx)
			data := fixtureTar(t,
				tarEntry{header: tar.Header{Name: ".", Typeflag: tar.TypeDir, Mode: mode}},
				tarEntry{header: tar.Header{Name: "nested", Typeflag: tar.TypeDir, Mode: mode}},
				regularEntry("nested/file", "contents"),
			)
			require.NoError(t, copyFixture(ctx, localFixture(t, fixtureLayer{title: ".", unpack: true, data: data}), dst))
			require.NoError(t, dst.Cleanup(ctx))
			_, err := os.Stat(dst.TmpDir())
			require.True(t, os.IsNotExist(err), "Cleanup must remove the job-owned directory: %v", err)
		})
	}
}

func TestExtractionReadsLegacyArtifactManifest(t *testing.T) {
	ctx := context.Background()
	source := localFixture(t, fixtureLayer{title: "nested/file", data: []byte("legacy artifact")})
	manifest, err := source.Resolve(ctx, defaultLocalTag)
	require.NoError(t, err)
	layers := manifestLayers(t, ctx, source, manifest)
	legacy, err := oras.Pack(ctx, source, defaultArtifactType, layers, oras.PackOptions{})
	require.NoError(t, err)
	require.Equal(t, "application/vnd.oci.artifact.manifest.v1+json", legacy.MediaType)
	require.NoError(t, source.Tag(ctx, legacy, defaultLocalTag))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, source, dst))
	data, err := afero.ReadFile(dst.FS(), "nested/file")
	require.NoError(t, err)
	require.Equal(t, "legacy artifact", string(data))
}

func TestArchiveLifecycle(t *testing.T) {
	ctx := context.Background()
	a := New()
	require.Nil(t, a.FS())
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	require.ErrorIs(t, a.Initialize(canceled), context.Canceled)
	require.Empty(t, a.TmpDir())
	require.NoError(t, a.Initialize(ctx))
	t.Cleanup(func() { require.NoError(t, a.Cleanup(ctx)) })
	require.ErrorContains(t, a.Initialize(ctx), "already initialized")
	_, err := a.FS().Stat(".")
	require.NoError(t, err, "double initialization must not close the original root")
	require.NoError(t, a.Cleanup(ctx))
	_, err = a.FS().Stat(".")
	require.Error(t, err)
	source := newTestSource(t, afero.NewOsFs(), t.TempDir())
	require.ErrorContains(t, a.Pack(ctx, zap.NewNop(), source, nil), "closed")
	require.ErrorContains(t, a.Unpack(ctx, nil, defaultLocalTag), "closed")
	require.NoError(t, a.Cleanup(ctx))
}

func TestCleanupDoesNotDeleteReplacementJobTree(t *testing.T) {
	ctx := context.Background()
	parent := t.TempDir()
	dst := newTestArchive(t, ctx)
	moved := filepath.Join(parent, "pinned-job")
	require.NoError(t, os.Rename(dst.TmpDir(), moved))
	require.NoError(t, os.Mkdir(dst.TmpDir(), 0700))
	canary := filepath.Join(dst.TmpDir(), "canary")
	t.Cleanup(func() {
		require.NoError(t, os.Remove(canary))
		require.NoError(t, os.Remove(dst.TmpDir()))
	})
	require.NoError(t, os.WriteFile(canary, []byte("replacement"), 0600))
	require.Error(t, dst.Cleanup(ctx), "the nonempty replacement directory must not be recursively removed")
	data, err := os.ReadFile(canary)
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
	entries, err := os.ReadDir(moved)
	require.NoError(t, err)
	require.Empty(t, entries, "cleanup must empty the pinned job directory instead")
}
