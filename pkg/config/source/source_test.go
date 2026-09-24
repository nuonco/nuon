package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
