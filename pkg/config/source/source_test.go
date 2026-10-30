package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

func TestReadSourceUsesBaseDir(t *testing.T) {
	baseDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "values.yaml"), []byte("base"), 0644))

	otherDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(otherDir, "values.yaml"), []byte("cwd"), 0644))
	t.Chdir(otherDir)

	contents, err := ReadSourceFrom("values.yaml", baseDir)
	require.NoError(t, err)
	require.Equal(t, "base", string(contents))
}

func TestReadSourceFallsBackToWorkingDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("cwd"), 0644))
	t.Chdir(dir)

	contents, err := ReadSource("values.yaml")
	require.NoError(t, err)
	require.Equal(t, "cwd", string(contents))
}

func TestSourceFromRepository(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	baseDir := filepath.Join(root, "apps", "app")
	require.NoError(t, os.MkdirAll(baseDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(root, "shared"), 0755))
	inside := filepath.Join(root, "shared", "config.toml")
	outside := filepath.Join(parent, "outside.toml")
	const contents = "name = 'trusted'\n"
	require.NoError(t, os.WriteFile(inside, []byte(contents), 0644))
	require.NoError(t, os.WriteFile(outside, []byte("name = 'canary'\n"), 0644))
	require.NoError(t, os.Symlink("../../shared/config.toml", filepath.Join(baseDir, "allowed.toml")))
	require.NoError(t, os.Symlink(inside, filepath.Join(baseDir, "absolute-link.toml")))
	require.NoError(t, os.Symlink(outside, filepath.Join(baseDir, "escape.toml")))
	require.NoError(t, os.Symlink("../../../outside.toml", filepath.Join(baseDir, "relative-escape.toml")))
	require.NoError(t, os.Symlink(parent, filepath.Join(baseDir, "escape-dir")))
	fs, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fs.Close()) })

	for _, path := range []string{"../../shared/config.toml", inside, "allowed.toml", "absolute-link.toml"} {
		t.Run("allowed/"+path, func(t *testing.T) {
			byts, err := ReadSourceFrom(path, baseDir, fs)
			require.NoError(t, err)
			require.Equal(t, contents, string(byts))
			obj, err := LoadSourceFrom(path, baseDir, fs)
			require.NoError(t, err)
			require.Equal(t, "trusted", obj["name"])
		})
	}
	for _, path := range []string{outside, "../../../outside.toml", "escape.toml", "relative-escape.toml", "escape-dir/outside.toml", root + "-outside/config.toml"} {
		t.Run("denied/"+path, func(t *testing.T) {
			byts, err := ReadSourceFrom(path, baseDir, fs)
			require.ErrorContains(t, err, "escapes repository")
			require.Nil(t, byts)
			obj, err := LoadSourceFrom(path, baseDir, fs)
			require.ErrorContains(t, err, "escapes repository")
			require.Nil(t, obj)
		})
	}

	byts, err := ReadSourceFrom("shared/config.toml", "", fs)
	require.NoError(t, err)
	require.Equal(t, contents, string(byts))
	for _, path := range []string{outside, "escape.toml"} {
		obj, err := LoadSourceFrom(path, baseDir, nil)
		require.NoError(t, err)
		require.Equal(t, "canary", obj["name"])
	}
}

func TestSourceFromBorrowedFilesystem(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("name = 'host'\n"), 0644))
	memory := afero.NewMemMapFs()
	require.NoError(t, memory.MkdirAll(root, 0755))
	require.NoError(t, afero.WriteFile(memory, path, []byte("name = 'borrowed'\n"), 0644))
	fs, err := sourcefs.From(afero.NewBasePathFs(memory, root), root)
	require.NoError(t, err)

	obj, err := LoadSourceFrom("config.toml", root, fs)
	require.NoError(t, err)
	require.Equal(t, "borrowed", obj["name"])
}

func TestSourceFromPinnedRepository(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	require.NoError(t, os.Mkdir(root, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config.toml"), []byte("name = 'pinned'\n"), 0644))
	fs, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fs.Close()) })
	require.NoError(t, os.Rename(root, root+"-original"))
	require.NoError(t, os.Mkdir(root, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config.toml"), []byte("name = 'replacement'\n"), 0644))

	obj, err := LoadSourceFrom("config.toml", root, fs)
	require.NoError(t, err)
	require.Equal(t, "pinned", obj["name"])
	require.NoError(t, fs.Close())
	obj, err = LoadSourceFrom("config.toml", root, fs)
	require.Error(t, err)
	require.Nil(t, obj)
}
