package dir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// newOsParser builds a parser rooted at dir on the real filesystem. These tests
// need real symlinks, which MemMapFs does not support.
func newOsParser(t *testing.T, dir string) *parser {
	t.Helper()

	return &parser{
		fs:   afero.Afero{Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)},
		opts: &ParseOptions{Ext: ".toml"},
	}
}

func TestListDir_FollowsSymlinkedDir(t *testing.T) {
	root := t.TempDir()

	// shared/components/images/bauleiter.toml, symlinked into the app dir as
	// app/components/images -- the layout that broke `nuon apps sync`.
	shared := filepath.Join(root, "shared", "components", "images")
	require.NoError(t, os.MkdirAll(shared, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(shared, "bauleiter.toml"), []byte("name = \"img_bauleiter\"\n"), 0o644))

	app := filepath.Join(root, "app")
	require.NoError(t, os.MkdirAll(filepath.Join(app, "components"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(app, "components", "0-vpc.toml"), []byte("name = \"vpc\"\n"), 0o644))
	require.NoError(t, os.Symlink(
		filepath.Join("..", "..", "shared", "components", "images"),
		filepath.Join(app, "components", "images"),
	))

	files, err := newOsParser(t, app).listDir("components")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		"components/0-vpc.toml",
		"components/images/bauleiter.toml",
	}, files)
}

func TestListDir_FollowsSymlinkedFile(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "components"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "shared.toml"), []byte("name = \"shared\"\n"), 0o644))
	require.NoError(t, os.Symlink(
		filepath.Join("..", "shared.toml"),
		filepath.Join(root, "components", "shared.toml"),
	))

	files, err := newOsParser(t, root).listDir("components")
	require.NoError(t, err)
	require.Equal(t, []string{"components/shared.toml"}, files)
}

func TestListDir_SkipsDanglingSymlink(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "components"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "components", "real.toml"), []byte("name = \"real\"\n"), 0o644))
	require.NoError(t, os.Symlink("./nowhere", filepath.Join(root, "components", "gone.toml")))
	require.NoError(t, os.Symlink("./nowhere-dir", filepath.Join(root, "components", "images")))

	files, err := newOsParser(t, root).listDir("components")
	require.NoError(t, err)
	require.Equal(t, []string{"components/real.toml"}, files)
}

func TestListDir_StopsOnSymlinkLoop(t *testing.T) {
	root := t.TempDir()

	// components/loop points back at components, so a naive walk never ends.
	components := filepath.Join(root, "components")
	require.NoError(t, os.MkdirAll(components, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(components, "real.toml"), []byte("name = \"real\"\n"), 0o644))
	require.NoError(t, os.Symlink("..", filepath.Join(components, "loop")))

	files, err := newOsParser(t, root).listDir("components")
	require.NoError(t, err)
	require.Equal(t, []string{"components/real.toml"}, files)
}

func TestListDir_MissingDirIsNotAnError(t *testing.T) {
	files, err := newOsParser(t, t.TempDir()).listDir("components")
	require.NoError(t, err)
	require.Empty(t, files)
}
