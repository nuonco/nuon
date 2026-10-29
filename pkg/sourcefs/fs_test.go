package sourcefs_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func TestRootedReads(t *testing.T) {
	for name, fs := range map[string]afero.Fs{
		"memory": afero.NewMemMapFs(),
		"os":     afero.NewOsFs(),
	} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			require.NoError(t, fs.MkdirAll(filepath.Join(root, "nested"), 0755))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "nested/input"), []byte("source"), 0755))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(filepath.Dir(root), "outside"), []byte("outside"), 0644))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			data, err := source.ReadFile("nested/input")
			require.NoError(t, err)
			require.Equal(t, "source", string(data))
			sub, err := source.Sub(".")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sub.Close()) })
			data, err = sub.ReadFile("nested/input")
			require.NoError(t, err)
			require.Equal(t, "source", string(data))
			file, err := source.Open("nested/input")
			require.NoError(t, err)
			data, err = io.ReadAll(file)
			require.NoError(t, err)
			require.NoError(t, file.Close())
			require.Equal(t, "source", string(data))
			info, err := source.Lstat("nested/input")
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0755), info.Mode().Perm())
			entries, err := source.ReadDir("nested")
			require.NoError(t, err)
			require.Len(t, entries, 1)
			require.Equal(t, "input", entries[0].Name())
			matches, err := source.Glob("nested/in*")
			require.NoError(t, err)
			require.Equal(t, []string{"nested/input"}, matches)
			var visited []string
			require.NoError(t, source.Walk("nested", func(name string, _ os.FileInfo, err error) error {
				visited = append(visited, name)
				return err
			}))
			require.Equal(t, []string{"nested", "nested/input"}, visited)
			for _, outside := range []string{"../outside", filepath.Join(filepath.Dir(root), "outside"), root + "-sibling/input"} {
				_, err := source.ReadFile(outside)
				require.Error(t, err)
				_, err = source.Open(outside)
				require.Error(t, err)
				_, err = source.ReadDir(outside)
				require.Error(t, err)
				_, err = source.Lstat(outside)
				require.Error(t, err)
				_, err = source.Glob(outside)
				require.Error(t, err)
				err = source.Walk(outside, func(string, os.FileInfo, error) error {
					t.Fatal("walk escaped the source root")
					return nil
				})
				require.Error(t, err)
			}
			nested, err := source.Sub("nested")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, nested.Close()) })
			_, err = nested.ReadFile("../nested/input")
			require.Error(t, err)
			_, err = sourcefs.New(fs, filepath.Join(root, "nested/input"))
			require.ErrorContains(t, err, "not a directory")
		})
	}
}

func TestRootedLinks(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(parent, "outside"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested/input"), []byte("source"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(parent, "outside/input"), []byte("outside"), 0644))
	canonical, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	for name, target := range map[string]string{
		"relative":    "nested",
		"absolute":    filepath.Join(canonical, "nested"),
		"escape":      "../outside",
		"prefix":      canonical + "-sibling",
		"return":      "../outside/../repo/nested",
		"cycle":       "cycle",
		"file-parent": "nested/input/../nested",
	} {
		require.NoError(t, os.Symlink(target, filepath.Join(root, name)))
	}
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	for _, name := range []string{"relative", "absolute"} {
		data, err := source.ReadFile(name + "/input")
		require.NoError(t, err)
		require.Equal(t, "source", string(data))
		resolved, err := source.Resolve(name + "/input")
		require.NoError(t, err)
		require.Equal(t, "nested/input", resolved)
		link, err := source.Readlink(name)
		require.NoError(t, err)
		require.NotEmpty(t, link)
	}
	for _, name := range []string{"escape", "prefix", "return", "cycle", "file-parent"} {
		_, err := source.Sub(name)
		require.Error(t, err, name)
		_, err = source.ReadFile(name + "/input")
		require.Error(t, err, name)
		_, err = source.Open(name + "/input")
		require.Error(t, err, name)
		_, err = source.ReadDir(name)
		require.Error(t, err, name)
		_, err = source.Lstat(name + "/input")
		require.Error(t, err, name)
		_, err = source.Readlink(name + "/input")
		require.Error(t, err, name)
	}
	require.NoError(t, os.Symlink(root, filepath.Join(parent, "checkout")))
	require.NoError(t, os.Symlink(filepath.Join(parent, "checkout/nested"), filepath.Join(root, "aliased")))
	relocated, err := sourcefs.New(afero.NewOsFs(), filepath.Join(parent, "checkout"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, relocated.Close()) })
	for _, name := range []string{"nested/input", "aliased/input", filepath.Join(parent, "checkout/nested/input")} {
		data, err := relocated.ReadFile(name)
		require.NoError(t, err)
		require.Equal(t, "source", string(data))
	}
}

func TestNewRejectsUnpinnedHostWrappers(t *testing.T) {
	root := t.TempDir()
	for name, tc := range map[string]struct {
		fs   afero.Fs
		root string
	}{
		"read-only": {afero.NewReadOnlyFs(afero.NewOsFs()), root},
		"base-path": {afero.NewBasePathFs(afero.NewOsFs(), root), "/"},
	} {
		t.Run(name, func(t *testing.T) {
			source, err := sourcefs.New(tc.fs, tc.root)
			if source != nil {
				t.Cleanup(func() { require.NoError(t, source.Close()) })
			}
			require.ErrorContains(t, err, "unsupported source filesystem")
		})
	}
}

func TestSubSymlinkedScope(t *testing.T) {
	root := t.TempDir()
	canonical, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(filepath.Join(root, "real"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "real/input"), []byte("scoped source"), 0o600))
	require.NoError(t, os.Symlink("../real", filepath.Join(root, "real/escape")))
	require.NoError(t, os.Symlink(canonical+"/real/../real", filepath.Join(root, "real/absolute-escape")))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	for name, target := range map[string]string{"relative": "real", "absolute": filepath.Join(canonical, "real")} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.Symlink(target, filepath.Join(root, name)))
			sub, err := source.Sub(name)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, sub.Close()) })
			resolved, err := sub.Resolve("input")
			require.NoError(t, err)
			require.Equal(t, "input", resolved)
			contents, err := sub.ReadFile(resolved)
			require.NoError(t, err)
			require.Equal(t, "scoped source", string(contents))
			for _, outside := range []string{"escape/input", "absolute-escape/input", "../real/input", root} {
				_, err := sub.Resolve(outside)
				require.Error(t, err, outside)
			}
		})
	}
}

func TestAferoBasePathNamespace(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/repo", 0755))
	require.NoError(t, afero.WriteFile(fs, "/repo/input", []byte("source"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/outside", []byte("outside"), 0644))
	source, err := sourcefs.From(afero.NewBasePathFs(fs, "/repo"), "/")
	require.NoError(t, err)
	data, err := source.ReadFile("input")
	require.NoError(t, err)
	require.Equal(t, "source", string(data))
	_, err = source.ReadFile("../outside")
	require.Error(t, err)
	_, err = source.ReadFile("../input")
	require.Error(t, err)
	_, err = source.ReadFile("/outside")
	require.Error(t, err)
}

func TestSourceRootCannotBeRedirectedAfterOpening(t *testing.T) {
	parent := t.TempDir()
	root, neighbor := filepath.Join(parent, "job"), filepath.Join(parent, "job-neighbor")
	require.NoError(t, os.MkdirAll(root, 0700))
	require.NoError(t, os.MkdirAll(neighbor, 0700))
	require.NoError(t, os.Mkdir(filepath.Join(root, "nested"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "input"), []byte("job source"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested/input"), []byte("nested job source"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(neighbor, "input"), []byte("neighbor canary"), 0600))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	require.NoError(t, os.Rename(root, root+"-moved"))
	require.NoError(t, os.Symlink(neighbor, root))
	data, err := source.ReadFile("input")
	require.NoError(t, err)
	require.Equal(t, "job source", string(data))
	file, err := source.Open("input")
	require.NoError(t, err)
	data, err = io.ReadAll(file)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.Equal(t, "job source", string(data))
	child, err := source.Sub(filepath.Join(root, "nested"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, child.Close()) })
	require.NoError(t, source.Close())
	data, err = child.ReadFile("input")
	require.NoError(t, err)
	require.Equal(t, "nested job source", string(data))
	_, err = child.ReadFile("../input")
	require.Error(t, err)
}

type swapBeforeOpenFS struct {
	*rootfs.FS
	swap func()
}

func (f *swapBeforeOpenFS) Open(name string) (afero.File, error) {
	f.swap()
	return f.FS.Open(name)
}

func TestSourceReadRejectsSymlinkSwapBeforeOpen(t *testing.T) {
	for _, swapDirectory := range []bool{false, true} {
		t.Run(fmt.Sprint("directory=", swapDirectory), func(t *testing.T) {
			parent := t.TempDir()
			root, neighbor := filepath.Join(parent, "job"), filepath.Join(parent, "neighbor")
			require.NoError(t, os.MkdirAll(filepath.Join(root, "data"), 0700))
			require.NoError(t, os.Mkdir(neighbor, 0700))
			require.NoError(t, os.WriteFile(filepath.Join(root, "data/input"), []byte("job source"), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "input"), []byte("neighbor canary"), 0600))
			pinned, err := rootfs.Open(root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, pinned.Close()) })
			fs := &swapBeforeOpenFS{FS: pinned, swap: func() {
				name, target := filepath.Join(root, "data/input"), filepath.Join(neighbor, "input")
				if swapDirectory {
					name, target = filepath.Join(root, "data"), neighbor
				}
				require.NoError(t, os.Rename(name, name+"-moved"))
				require.NoError(t, os.Symlink(target, name))
			}}
			source, err := sourcefs.From(fs, root)
			require.NoError(t, err)
			_, err = source.ReadFile("data/input")
			require.Error(t, err)
			canary, err := os.ReadFile(filepath.Join(neighbor, "input"))
			require.NoError(t, err)
			require.Equal(t, "neighbor canary", string(canary))
		})
	}
}

func TestGitMetadataNotEnumerable(t *testing.T) {
	for _, gitDir := range []string{".git", ".GIT"} {
		t.Run(gitDir, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(root, gitDir), 0700))
			require.NoError(t, os.WriteFile(filepath.Join(root, gitDir, "HEAD"), []byte("private metadata"), 0600))
			require.NoError(t, os.Mkdir(filepath.Join(root, ".config"), 0700))
			require.NoError(t, os.WriteFile(filepath.Join(root, ".config", "input"), []byte("authored source"), 0600))
			require.NoError(t, os.Symlink(gitDir, filepath.Join(root, "alias")))
			require.NoError(t, os.Symlink("alias/HEAD", filepath.Join(root, "file-alias")))
			pinned, err := rootfs.Open(root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, pinned.Close()) })
			source, err := sourcefs.From(pinned, pinned.Root())
			require.NoError(t, err)
			for _, name := range []string{gitDir, gitDir + "/HEAD", "alias", "alias/HEAD", "file-alias"} {
				_, err := source.ReadFile(name)
				require.ErrorContains(t, err, "Git metadata", name)
				_, err = source.Lstat(name)
				require.ErrorContains(t, err, "Git metadata", name)
				_, err = source.Readlink(name)
				require.ErrorContains(t, err, "Git metadata", name)
			}
			entries, err := source.ReadDir(".")
			require.NoError(t, err)
			require.Len(t, entries, 1)
			require.Equal(t, ".config", entries[0].Name())
			for _, count := range []int{-1, 1} {
				directory, err := source.Open(".")
				require.NoError(t, err)
				entries, err := directory.Readdir(count)
				require.NoError(t, err)
				require.Len(t, entries, 1)
				require.Equal(t, ".config", entries[0].Name())
				entries, err = directory.Readdir(1)
				require.Empty(t, entries)
				require.ErrorIs(t, err, io.EOF)
				require.NoError(t, directory.Close())
				directory, err = source.Open(".")
				require.NoError(t, err)
				names, err := directory.Readdirnames(count)
				require.NoError(t, err)
				require.Equal(t, []string{".config"}, names)
				names, err = directory.Readdirnames(1)
				require.Empty(t, names)
				require.ErrorIs(t, err, io.EOF)
				require.NoError(t, directory.Close())
			}
			for _, pattern := range []string{"*", "*/*", "alias/*"} {
				matches, err := source.Glob(pattern)
				require.NoError(t, err)
				switch pattern {
				case "*":
					require.Equal(t, []string{".config"}, matches)
				case "*/*":
					require.Equal(t, []string{filepath.Join(".config", "input")}, matches)
				default:
					require.Empty(t, matches)
				}
			}
			var visited []string
			require.NoError(t, source.Walk(".", func(name string, _ os.FileInfo, err error) error {
				visited = append(visited, name)
				return err
			}))
			require.Equal(t, []string{".", ".config", filepath.Join(".config", "input")}, visited)
			data, err := source.ReadFile(".config/input")
			require.NoError(t, err)
			require.Equal(t, "authored source", string(data))
		})
	}
}
