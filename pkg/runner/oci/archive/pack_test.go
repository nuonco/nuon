package ociarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

func TestPackMemoryMatchesDisk(t *testing.T) {
	ctx := context.Background()
	var packedDigest string
	for _, backend := range []string{"memory", "disk"} {
		t.Run(backend, func(t *testing.T) {
			var fs afero.Fs = afero.NewMemMapFs()
			root := "/source"
			if backend == "disk" {
				fs, root = afero.NewOsFs(), t.TempDir()
			}
			require.NoError(t, fs.MkdirAll(root, 0755))
			source := newTestSource(t, fs, root)
			root = source.Root()
			files := []FileRef{
				writeInputFile(t, fs, root, "nested/main.tf", "terraform {}", 0600),
				writeInputFile(t, fs, root, "bin/terraform", "executable", 0750),
				writeInputFile(t, fs, root, "empty.txt", "", 0644),
			}
			writeInputFile(t, fs, root, ".git/config", "ignored", 0644)
			writeInputFile(t, fs, root, "nested/.git/HEAD", "ignored", 0644)
			enumerated, err := SourceFiles(ctx, source, "application/octet-stream")
			require.NoError(t, err)
			require.ElementsMatch(t, files, enumerated)
			a := newTestArchive(t, ctx)
			require.NoError(t, a.Pack(ctx, zap.NewNop(), source, enumerated))
			dgst := resolveTestManifest(t, ctx, a).Digest.String()
			if packedDigest == "" {
				packedDigest = dgst
			} else {
				require.Equal(t, packedDigest, dgst)
			}
			dst := newTestArchive(t, ctx)
			require.NoError(t, copyFixture(ctx, a.Ref(), dst))
			for _, input := range files {
				got, err := afero.ReadFile(dst.FS(), input.RelPath)
				require.NoError(t, err)
				want, err := source.ReadFile(input.AbsPath)
				require.NoError(t, err)
				require.Equal(t, want, got)
			}
			info, err := dst.FS().Stat("bin/terraform")
			require.NoError(t, err)
			require.EqualValues(t, 0755, info.Mode().Perm())
		})
	}
}

func TestPackReadsOnlySuppliedSource(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	host := writeTestFile(t, root, "file", "host canary", 0644)
	fs := afero.NewMemMapFs()
	input := writeInputFile(t, fs, root, "file", "memory input", 0755)
	source := newTestSource(t, fs, root)
	packed := newTestArchive(t, ctx)
	require.NoError(t, packed.Pack(ctx, zap.NewNop(), source, []FileRef{input}))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, packed.Ref(), dst))
	data, err := afero.ReadFile(dst.FS(), "file")
	require.NoError(t, err)
	require.Equal(t, "memory input", string(data))
	require.NoError(t, fs.Remove(input.AbsPath))
	missing := newTestArchive(t, ctx)
	require.Error(t, missing.Pack(ctx, zap.NewNop(), source, []FileRef{host}))
	_, err = missing.Ref().Resolve(ctx, defaultLocalTag)
	require.Error(t, err, "a host-only input must not publish a manifest")
	data, err = os.ReadFile(host.AbsPath)
	require.NoError(t, err)
	require.Equal(t, "host canary", string(data))
}

func TestPackRejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	for _, backend := range []string{"memory", "disk"} {
		t.Run(backend, func(t *testing.T) {
			var fs afero.Fs = afero.NewMemMapFs()
			root := "/source"
			if backend == "disk" {
				fs, root = afero.NewOsFs(), t.TempDir()
			}
			input := writeInputFile(t, fs, root, "file", "input", 0644)
			source := newTestSource(t, fs, root)
			for _, name := range []string{"", ".", "/absolute", "../escape", "dir/../../escape", "dir/../file", "./file", `..\escape`, "C:/escape", "nul\x00"} {
				t.Run(fmt.Sprintf("output=%q", name), func(t *testing.T) {
					a := newTestArchive(t, ctx)
					ref := input
					ref.RelPath = name
					require.Error(t, a.Pack(ctx, zap.NewNop(), source, []FileRef{ref}))
					_, err := a.Ref().Resolve(ctx, defaultLocalTag)
					require.Error(t, err)
				})
			}
			for _, name := range []string{filepath.Join(filepath.Dir(root), "outside", "file"), "../outside/file"} {
				a := newTestArchive(t, ctx)
				ref := input
				ref.AbsPath = name
				require.ErrorContains(t, a.Pack(ctx, zap.NewNop(), source, []FileRef{ref}), "escapes repository")
				_, err := a.Ref().Resolve(ctx, defaultLocalTag)
				require.Error(t, err)
			}
			duplicate := newTestArchive(t, ctx)
			require.ErrorContains(t, duplicate.Pack(ctx, zap.NewNop(), source, []FileRef{input, input}), "duplicate archive path")
			require.ErrorContains(t, duplicate.Pack(ctx, zap.NewNop(), nil, nil), "source filesystem is required")
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			_, err := SourceFiles(canceled, source, "")
			require.ErrorIs(t, err, context.Canceled)
			require.ErrorIs(t, newTestArchive(t, ctx).Pack(canceled, zap.NewNop(), source, []FileRef{input}), context.Canceled)
		})
	}
}

func TestPackRootRelativeInputAndRenamedOutput(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	input := writeInputFile(t, fs, "/source", "generated/chart-1.0.tgz", "chart", 0644)
	source := newTestSource(t, fs, "/source")
	input.RelPath = "chart.tgz"
	absolute := newTestArchive(t, ctx)
	require.NoError(t, absolute.Pack(ctx, zap.NewNop(), source, []FileRef{input}))
	input.AbsPath = "generated/chart-1.0.tgz"
	relative := newTestArchive(t, ctx)
	require.NoError(t, relative.Pack(ctx, zap.NewNop(), source, []FileRef{input}))
	require.Equal(t, resolveTestManifest(t, ctx, absolute).Digest, resolveTestManifest(t, ctx, relative).Digest)
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, relative.Ref(), dst))
	data, err := afero.ReadFile(dst.FS(), "chart.tgz")
	require.NoError(t, err)
	require.Equal(t, "chart", string(data))
}

func TestPackOSFileAndDirectoryLinks(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeTestFile(t, root, "target/main.tf", "safe", 0750)
	require.NoError(t, os.Mkdir(filepath.Join(root, "nested"), 0755))
	require.NoError(t, os.Symlink("../target", filepath.Join(root, "nested/link")))
	require.NoError(t, os.Symlink("target/main.tf", filepath.Join(root, "relative")))
	require.NoError(t, os.Symlink(filepath.Join(root, "target/main.tf"), filepath.Join(root, "absolute")))
	source := newTestSource(t, afero.NewOsFs(), root)
	files, err := SourceFiles(ctx, source, "application/octet-stream")
	require.NoError(t, err)
	require.Len(t, files, 4, "directory links must not be followed")
	packed := newTestArchive(t, ctx)
	require.NoError(t, packed.Pack(ctx, zap.NewNop(), source, files))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, packed.Ref(), dst))
	for _, name := range []string{"nested/link/main.tf", "relative", "absolute"} {
		data, err := afero.ReadFile(dst.FS(), name)
		require.NoError(t, err)
		require.Equal(t, "safe", string(data))
		info, err := dst.FS().Stat(name)
		require.NoError(t, err)
		require.EqualValues(t, 0755, info.Mode().Perm())
	}
	info, _, err := dst.FS().(afero.Lstater).LstatIfPossible("nested/link")
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink)
}

func TestPackOSRejectsEscapingLinks(t *testing.T) {
	for _, link := range []string{"../outside/secret", "../outside", "$outside/secret", "$outside", "link", "missing"} {
		t.Run(link, func(t *testing.T) {
			ctx := context.Background()
			parent := t.TempDir()
			root, outside := filepath.Join(parent, "source"), filepath.Join(parent, "outside")
			writeTestFile(t, root, "inside", "safe", 0644)
			canary := writeTestFile(t, outside, "secret", "not an input", 0644)
			require.NoError(t, os.Symlink(strings.ReplaceAll(link, "$outside", outside), filepath.Join(root, "link")))
			source := newTestSource(t, afero.NewOsFs(), root)
			_, err := SourceFiles(ctx, source, "")
			require.Error(t, err)
			for _, name := range []string{"link", "link/secret"} {
				a := newTestArchive(t, ctx)
				require.Error(t, a.Pack(ctx, zap.NewNop(), source, []FileRef{{AbsPath: name, RelPath: "secret"}}))
				_, err = a.Ref().Resolve(ctx, defaultLocalTag)
				require.Error(t, err)
			}
			data, err := os.ReadFile(canary.AbsPath)
			require.NoError(t, err)
			require.Equal(t, "not an input", string(data))
		})
	}
}

func TestPackedLayersRemainReadableByLegacyORAS(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewOsFs()
	root := t.TempDir()
	inputs := []FileRef{writeInputFile(t, fs, root, "nested/main.tf", "terraform {}", 0644), writeInputFile(t, fs, root, "tool", "executable", 0755)}
	packed := newTestArchive(t, ctx)
	require.NoError(t, packed.Pack(ctx, zap.NewNop(), newTestSource(t, fs, root), inputs))
	out := t.TempDir()
	legacy, err := file.New(out)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, legacy.Close()) })
	_, err = oras.Copy(ctx, packed.Ref(), defaultLocalTag, legacy, defaultLocalTag, oras.DefaultCopyOptions)
	require.NoError(t, err)
	for _, input := range inputs {
		contents, err := os.ReadFile(filepath.Join(out, input.RelPath))
		require.NoError(t, err)
		expected, err := os.ReadFile(input.AbsPath)
		require.NoError(t, err)
		require.Equal(t, expected, contents)
	}
	info, err := os.Stat(filepath.Join(out, "tool"))
	require.NoError(t, err)
	require.EqualValues(t, 0755, info.Mode().Perm())
}

func TestPackDirectoryLinkToSourceRoot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeTestFile(t, root, "file", "input", 0644)
	require.NoError(t, os.Symlink(".", filepath.Join(root, "root-link")))
	source := newTestSource(t, afero.NewOsFs(), root)
	files, err := SourceFiles(ctx, source, "application/octet-stream")
	require.NoError(t, err)
	require.Len(t, files, 2)
	packed := newTestArchive(t, ctx)
	require.NoError(t, packed.Pack(ctx, zap.NewNop(), source, files))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, packed.Ref(), dst))
	data, err := afero.ReadFile(dst.FS(), "root-link/file")
	require.NoError(t, err)
	require.Equal(t, "input", string(data))
}
