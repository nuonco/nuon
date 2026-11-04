package ociarchive

import (
	"archive/tar"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

func TestPackDirectoryLinkCompatibility(t *testing.T) {
	for _, fixture := range []struct {
		name  string
		links map[string]string
	}{
		{"dot", map[string]string{"a": "./real"}},
		{"dot-components", map[string]string{"a": "real/.././real/"}},
		{"chain", map[string]string{"a": "real", "b": "a"}},
		{"forward-chain", map[string]string{"a": "z", "z": "./real"}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			writeTestFile(t, root, "real/file", "executable", 0755)
			for name, target := range fixture.links {
				require.NoError(t, os.Symlink(target, filepath.Join(root, name)))
			}
			source := newTestSource(t, afero.NewOsFs(), root)
			inputs, err := SourceFiles(ctx, source, ".", "application/octet-stream")
			require.NoError(t, err)
			packed := newTestArchive(t, ctx)
			require.NoError(t, packed.Pack(ctx, zap.NewNop(), source, inputs))

			legacyRoot := t.TempDir()
			legacy, err := file.New(legacyRoot)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, legacy.Close()) })
			_, err = oras.Copy(ctx, packed.Ref(), defaultLocalTag, legacy, defaultLocalTag, oras.DefaultCopyOptions)
			require.NoError(t, err)
			assertDirectoryLinks(t, afero.NewBasePathFs(afero.NewOsFs(), legacyRoot), fixture.links, "executable")

			dst := newTestArchive(t, ctx)
			require.NoError(t, copyFixture(ctx, packed.Ref(), dst))
			assertDirectoryLinks(t, dst.FS(), fixture.links, "executable")
		})
	}
}

func TestPackDirectoryLinksToOmittedEmptyDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "empty"), 0755))
	require.NoError(t, os.Symlink("./empty", filepath.Join(root, "a")))
	require.NoError(t, os.Symlink("a", filepath.Join(root, "b")))
	source := newTestSource(t, afero.NewOsFs(), root)
	inputs, err := SourceFiles(ctx, source, ".", "application/octet-stream")
	require.NoError(t, err)
	require.Len(t, inputs, 2)
	packed := newTestArchive(t, ctx)
	require.NoError(t, packed.Pack(ctx, zap.NewNop(), source, inputs))
	dst := newTestArchive(t, ctx)
	require.NoError(t, copyFixture(ctx, packed.Ref(), dst))
	for name, expected := range map[string]string{"a": "./empty", "b": "a"} {
		target, err := dst.FS().(afero.LinkReader).ReadlinkIfPossible(name)
		require.NoError(t, err)
		require.Equal(t, expected, target)
	}
	_, err = dst.FS().Stat("empty")
	require.True(t, os.IsNotExist(err), "empty directories are omitted by legacy packaging")
}

func TestExtractionReadsFrozenLegacyDirectoryLinks(t *testing.T) {
	for _, fixture := range []struct {
		name  string
		links map[string]string
	}{
		{"legacy-dot-link.tar.gz", map[string]string{"a": "./real"}},
		{"legacy-chain-links.tar.gz", map[string]string{"a": "real", "b": "a"}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			ctx := context.Background()
			data, err := os.ReadFile(filepath.Join("testdata", fixture.name))
			require.NoError(t, err)
			dst := newTestArchive(t, ctx)
			require.NoError(t, copyFixture(ctx, localFixture(t, fixtureLayer{title: ".", unpack: true, data: data}), dst))
			assertDirectoryLinks(t, dst.FS(), fixture.links, "legacy executable\n")
		})
	}
}

func assertDirectoryLinks(t *testing.T, fs afero.Fs, links map[string]string, contents string) {
	t.Helper()
	for name, target := range links {
		actual, err := fs.(afero.LinkReader).ReadlinkIfPossible(name)
		require.NoError(t, err)
		require.Equal(t, target, actual, "link targets must be preserved verbatim")
		data, err := afero.ReadFile(fs, name+"/file")
		require.NoError(t, err)
		require.Equal(t, contents, string(data))
		info, err := fs.Stat(name + "/file")
		require.NoError(t, err)
		require.EqualValues(t, 0755, info.Mode().Perm())
	}
}

func TestExtractionResolvesLinksAfterAllLayers(t *testing.T) {
	ctx := context.Background()
	dst := newTestArchive(t, ctx)
	source := localFixture(t,
		fixtureLayer{title: ".", unpack: true, data: fixtureTar(t,
			linkEntry("a", "b", tar.TypeSymlink),
			linkEntry("missing", "omitted-empty-directory", tar.TypeSymlink),
			linkEntry("missing-chain", "missing", tar.TypeSymlink),
		)},
		fixtureLayer{title: ".", unpack: true, data: fixtureTar(t,
			linkEntry("b", "real", tar.TypeSymlink),
			linkEntry("root", "./real/..", tar.TypeSymlink),
		)},
		fixtureLayer{title: "real/file", data: []byte("safe")},
	)
	require.NoError(t, copyFixture(ctx, source, dst))
	for _, name := range []string{"a/file", "b/file", "root/real/file"} {
		data, err := afero.ReadFile(dst.FS(), name)
		require.NoError(t, err)
		require.Equal(t, "safe", string(data))
	}
	for name, expected := range map[string]string{"missing": "omitted-empty-directory", "missing-chain": "missing"} {
		target, err := dst.FS().(afero.LinkReader).ReadlinkIfPossible(name)
		require.NoError(t, err)
		require.Equal(t, expected, target)
	}
}

func TestExtractionRejectsCrossLayerLinkAttacks(t *testing.T) {
	for _, fixture := range []string{"escape", "loop", "write"} {
		t.Run(fixture, func(t *testing.T) {
			ctx := context.Background()
			dst := newTestArchive(t, ctx)
			canary := filepath.Join(dst.TmpDir(), "canary")
			require.NoError(t, os.WriteFile(canary, []byte("untouched"), 0600))
			first := fixtureLayer{title: ".", unpack: true}
			second := fixtureLayer{title: ".", unpack: true}
			switch fixture {
			case "escape":
				first.data = fixtureTar(t, linkEntry("a", "b/../canary", tar.TypeSymlink))
				second.data = fixtureTar(t, linkEntry("b", ".", tar.TypeSymlink))
			case "loop":
				first.data = fixtureTar(t, linkEntry("a", "./b", tar.TypeSymlink))
				second.data = fixtureTar(t, linkEntry("b", "./a", tar.TypeSymlink))
			case "write":
				first.data = fixtureTar(t, linkEntry("a", ".", tar.TypeSymlink), linkEntry("b", "a", tar.TypeSymlink))
				second = fixtureLayer{title: "b/file", data: []byte("bad")}
			}
			source := localFixture(t, first, second)
			require.Error(t, copyFixture(ctx, source, dst))
			require.Error(t, copyFixture(ctx, source, dst), "retries must not publish a rejected manifest")
			_, err := dst.Ref().Resolve(ctx, defaultLocalTag)
			require.Error(t, err)
			_, err = dst.FS().Stat(".")
			require.Error(t, err)
			data, err := os.ReadFile(canary)
			require.NoError(t, err)
			require.Equal(t, "untouched", string(data))
			_, err = os.Stat(filepath.Join(dst.BasePath(), "file"))
			require.True(t, os.IsNotExist(err), "no file may be written through a symlink chain")
		})
	}
}

func TestExtractionBoundsLinkResolution(t *testing.T) {
	for _, depth := range []int{8, 256} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			ctx := context.Background()
			dst := newTestArchive(t, ctx)
			entries := []tarEntry{regularEntry("real/file", "safe")}
			for i := 0; i < depth; i++ {
				target := "real"
				if i < depth-1 {
					target = fmt.Sprintf("link-%03d", i+1)
				}
				entries = append(entries, linkEntry(fmt.Sprintf("link-%03d", i), target, tar.TypeSymlink))
			}
			err := copyFixture(ctx, localFixture(t, fixtureLayer{title: ".", unpack: true, data: fixtureTar(t, entries...)}), dst)
			if depth == 256 {
				require.ErrorContains(t, err, "link depth")
				_, err = dst.Ref().Resolve(ctx, defaultLocalTag)
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				data, err := afero.ReadFile(dst.FS(), "link-000/file")
				require.NoError(t, err)
				require.Equal(t, "safe", string(data))
			}
		})
	}
}
