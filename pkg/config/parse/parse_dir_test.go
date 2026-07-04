package parse

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestParseDirErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := ParseDir(ctx, ParseConfig{
			Dirname: "/tmp/nonexistent-dir-abc123xyz",
		})
		require.Error(t, err)
		var errCfg config.ErrConfig
		require.ErrorAs(t, err, &errCfg)
		require.Contains(t, errCfg.Description, "does not exist")
	})

	t.Run("empty directory no toml files", func(t *testing.T) {
		dir := t.TempDir()

		_, err := ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to parse directory")
	})

	t.Run("invalid toml syntax", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "nuon.toml"), []byte("[[[bad toml"), 0644)
		require.NoError(t, err)

		_, err = ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to parse directory")
	})

	t.Run("valid toml missing version", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "nuon.toml"), []byte(`
[sandbox]
type = "unknown_type"
`), 0644)
		require.NoError(t, err)

		_, err = ParseDir(ctx, ParseConfig{
			Dirname: dir,
		})
		require.Error(t, err)
	})
}
