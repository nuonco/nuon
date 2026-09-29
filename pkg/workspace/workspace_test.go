package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWithCleanup(t *testing.T) {
	v := validator.New()
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("WithCleanup(true) removes existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-true"

		ws1, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws1.Init(ctx))

		testFile := filepath.Join(ws1.Root(), "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0o644)
		require.NoError(t, err)
		require.FileExists(t, testFile)

		ws2, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)
		require.NoError(t, ws2.Init(ctx))

		assert.NoFileExists(t, testFile)
		assert.DirExists(t, ws2.Root())
	})

	t.Run("WithCleanup(false) fails with existing directory and git source", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-false"

		ws1, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws1.Init(ctx))

		testFile := filepath.Join(ws1.Root(), "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0o644)
		require.NoError(t, err)

		ws2, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(false),
			WithGitSource(&GitSource{
				URL: "https://github.com/example/repo.git",
				Ref: "main",
			}),
		)
		require.NoError(t, err)

		err = ws2.Init(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to clone repo")
	})

	t.Run("WithCleanup(true) works with non-existent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-new"

		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)
		require.NoError(t, ws.Init(ctx))

		assert.DirExists(t, ws.Root())
	})

	t.Run("Default behavior without WithCleanup option", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-default"

		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws.Init(ctx))

		assert.False(t, ws.cleanupBeforeInit)
	})

	t.Run("WithCleanup(true) handles non-existent directory gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-nonexistent"

		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)

		err = ws.Init(ctx)
		require.NoError(t, err)

		assert.DirExists(t, ws.Root())
	})
}

func TestCleanupExistingDir(t *testing.T) {
	v := validator.New()
	logger, _ := zap.NewDevelopment()

	t.Run("cleanupExistingDir removes directory with contents", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-method"

		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)

		rootDir := ws.rootDir()
		err = os.MkdirAll(rootDir, defaultDirPermissions)
		require.NoError(t, err)

		testFile := filepath.Join(rootDir, "test.txt")
		err = os.WriteFile(testFile, []byte("content"), 0o644)
		require.NoError(t, err)

		subDir := filepath.Join(rootDir, "subdir")
		err = os.MkdirAll(subDir, defaultDirPermissions)
		require.NoError(t, err)

		err = ws.cleanupExistingDir()
		require.NoError(t, err)

		assert.NoDirExists(t, rootDir)
	})

	t.Run("cleanupExistingDir succeeds when directory doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-nodir"

		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)

		err = ws.cleanupExistingDir()
		require.NoError(t, err)
	})
}
