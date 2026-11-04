package ociarchive

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
)

func snapshotContentTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := make(map[string]string)
	require.NoError(t, filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			snapshot[rel] = "directory"
			return nil
		}
		data, err := os.ReadFile(name)
		snapshot[rel] = string(data)
		return err
	}))
	return snapshot
}

func TestContentStoreUsesPinnedRootsAfterReplacement(t *testing.T) {
	for _, scope := range []string{"job root", "content root"} {
		t.Run(scope, func(t *testing.T) {
			ctx := context.Background()
			a := newTestArchive(t, ctx)
			target := a.Ref().(oras.Target)
			seed := content.NewDescriptorFromBytes("test/blob", []byte("original"))
			require.NoError(t, target.Push(ctx, seed, strings.NewReader("original")))
			outside := t.TempDir()
			casPath := outside
			replaced := filepath.Join(a.TmpDir(), "content")
			if scope == "job root" {
				casPath = filepath.Join(outside, "content")
				replaced = a.TmpDir()
			}
			require.NoError(t, os.MkdirAll(filepath.Join(casPath, "blobs/sha256"), 0700))
			require.NoError(t, os.MkdirAll(filepath.Join(casPath, "ingest"), 0700))
			require.NoError(t, os.WriteFile(filepath.Join(casPath, "blobs/sha256", seed.Digest.Encoded()), []byte("neighbor"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(casPath, "index.json"), []byte("index canary"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(casPath, "ingest/canary"), []byte("ingest canary"), 0600))
			before := snapshotContentTree(t, outside)
			moved := filepath.Join(t.TempDir(), "held")
			require.NoError(t, os.Rename(replaced, moved))
			require.NoError(t, os.Symlink(outside, replaced))
			data, err := content.FetchAll(ctx, target, seed)
			require.NoError(t, err)
			require.Equal(t, "original", string(data))
			source := localFixture(t, fixtureLayer{title: "file", data: []byte("copied after replacement")})
			require.NoError(t, copyFixture(ctx, source, a))
			data, err = afero.ReadFile(a.FS(), "file")
			require.NoError(t, err)
			require.Equal(t, "copied after replacement", string(data))
			require.Equal(t, before, snapshotContentTree(t, outside))
			require.NoError(t, a.Cleanup(ctx))
			require.Equal(t, before, snapshotContentTree(t, outside), "cleanup must not walk a replacement tree")
			entries, err := os.ReadDir(moved)
			require.NoError(t, err)
			require.Empty(t, entries, "cleanup must empty even a CAS moved outside the job directory")
		})
	}
}

func TestContentStoreRejectsPreexistingSymlinks(t *testing.T) {
	for _, name := range []string{"blobs", "blobs/sha256", "ingest", "blob"} {
		for _, internal := range []bool{false, true} {
			t.Run(name+map[bool]string{false: "/external", true: "/internal"}[internal], func(t *testing.T) {
				ctx := context.Background()
				a := newTestArchive(t, ctx)
				store := a.Ref().(oras.Target)
				desc := content.NewDescriptorFromBytes("test/blob", []byte("original"))
				casRoot := filepath.Join(a.TmpDir(), "content")
				outside := t.TempDir()
				target := outside
				if internal {
					target = filepath.Join(casRoot, "neighbor")
					require.NoError(t, os.Mkdir(target, 0700))
				}
				require.NoError(t, os.MkdirAll(filepath.Join(target, "sha256"), 0700))
				require.NoError(t, os.WriteFile(filepath.Join(target, "canary"), []byte("neighbor"), 0600))
				before := snapshotContentTree(t, target)
				linkPath := filepath.Join(casRoot, name)
				linkTarget := target
				if name == "blob" {
					linkPath = filepath.Join(casRoot, "blobs/sha256", desc.Digest.Encoded())
					linkTarget = filepath.Join(target, "canary")
				} else {
					require.NoError(t, os.RemoveAll(linkPath))
				}
				if internal {
					var err error
					linkTarget, err = filepath.Rel(filepath.Dir(linkPath), linkTarget)
					require.NoError(t, err)
				}
				require.NoError(t, os.Symlink(linkTarget, linkPath))
				require.ErrorContains(t, store.Push(ctx, desc, strings.NewReader("original")), "symlinks are not allowed")
				if name != "ingest" {
					_, err := store.Fetch(ctx, desc)
					require.ErrorContains(t, err, "symlinks are not allowed")
					_, err = store.Exists(ctx, desc)
					require.ErrorContains(t, err, "symlinks are not allowed")
				}
				require.Equal(t, before, snapshotContentTree(t, target))
				if !internal {
					require.NoError(t, a.Cleanup(ctx))
					require.Equal(t, before, snapshotContentTree(t, target))
				}
			})
		}
	}
}

func TestContentStoreDoesNotOverwritePreexistingHardlink(t *testing.T) {
	ctx := context.Background()
	a := newTestArchive(t, ctx)
	store := a.Ref().(oras.Target)
	desc := content.NewDescriptorFromBytes("test/blob", []byte("original"))
	canary := filepath.Join(t.TempDir(), "canary")
	require.NoError(t, os.WriteFile(canary, []byte("neighbor"), 0600))
	require.NoError(t, os.Link(canary, filepath.Join(a.TmpDir(), "content/blobs/sha256", desc.Digest.Encoded())))
	require.ErrorIs(t, store.Push(ctx, desc, strings.NewReader("original")), errdef.ErrAlreadyExists)
	rc, err := store.Fetch(ctx, desc)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, rc)
	require.ErrorIs(t, err, content.ErrMismatchedDigest)
	require.NoError(t, rc.Close())
	require.NoError(t, a.Cleanup(ctx))
	data, err := os.ReadFile(canary)
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(data))
}

func TestContentStoreCorruptionCannotPublishExtractedArchive(t *testing.T) {
	ctx := context.Background()
	a := newTestArchive(t, ctx)
	store := a.Ref().(oras.Target)
	layer := content.NewDescriptorFromBytes("application/octet-stream", []byte("original"))
	layer.Annotations = map[string]string{v1.AnnotationTitle: "file"}
	require.NoError(t, store.Push(ctx, layer, strings.NewReader("original")))
	manifest, err := oras.Pack(ctx, store, defaultArtifactType, []v1.Descriptor{layer}, oras.PackOptions{PackImageManifest: true})
	require.NoError(t, err)
	// Tamper with a verified blob between ingestion and materialization.
	require.NoError(t, os.WriteFile(filepath.Join(a.TmpDir(), "content/blobs/sha256", layer.Digest.Encoded()), []byte("tampered"), 0600))
	require.ErrorIs(t, store.Tag(ctx, manifest, "published"), content.ErrMismatchedDigest)
	_, err = store.Resolve(ctx, "published")
	require.ErrorIs(t, err, errdef.ErrNotFound)
	_, err = a.FS().Open("file")
	require.Error(t, err, "rejected extraction must close its output capability")
}

func TestPackUsesPinnedSourceAndWorkspaceAfterReplacement(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	input := writeTestFile(t, root, "file", "source input", 0755)
	source := newTestSource(t, afero.NewOsFs(), root)
	a := newTestArchive(t, ctx)
	outside := t.TempDir()
	writeTestFile(t, outside, "file", "neighbor input", 0644)
	before := snapshotContentTree(t, outside)
	moved := filepath.Join(t.TempDir(), "source")
	require.NoError(t, os.Rename(root, moved))
	require.NoError(t, os.Symlink(outside, root))
	require.NoError(t, os.Rename(a.TmpDir(), filepath.Join(t.TempDir(), "archive")))
	require.NoError(t, os.Symlink(outside, a.TmpDir()))
	require.NoError(t, a.Pack(ctx, zap.NewNop(), source, []FileRef{input}))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, a.Ref(), dst))
	data, err := afero.ReadFile(dst.FS(), "file")
	require.NoError(t, err)
	require.Equal(t, "source input", string(data))
	require.Equal(t, before, snapshotContentTree(t, outside))
	require.NoError(t, a.Cleanup(ctx))
	require.Equal(t, before, snapshotContentTree(t, outside))
}
