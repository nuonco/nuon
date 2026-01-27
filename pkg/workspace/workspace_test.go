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
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-true"

		// Create workspace without cleanup first
		ws1, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws1.Init(ctx))

		// Create a file in the workspace to verify cleanup
		testFile := filepath.Join(ws1.Root(), "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0o644)
		require.NoError(t, err)
		require.FileExists(t, testFile)

		// Create new workspace with same ID and cleanup enabled
		ws2, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)
		require.NoError(t, ws2.Init(ctx))

		// Verify the old file is gone (directory was cleaned up)
		assert.NoFileExists(t, testFile)
		// Verify the directory exists again (it was re-created)
		assert.DirExists(t, ws2.Root())
	})

	t.Run("WithCleanup(false) fails with existing directory and git source", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-false"

		// Create workspace without git source first (this will succeed)
		ws1, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws1.Init(ctx))

		// Create a file in the workspace
		testFile := filepath.Join(ws1.Root(), "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0o644)
		require.NoError(t, err)

		// Try to create new workspace with same ID, git source, and no cleanup
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

		// Init should fail because directory exists and git clone requires empty dir
		err = ws2.Init(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to clone repo")
	})

	t.Run("WithCleanup(true) works with non-existent directory", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		workspaceID := "test-cleanup-new"

		// Create workspace with cleanup on a fresh directory
		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)
		require.NoError(t, ws.Init(ctx))

		// Verify the directory was created
		assert.DirExists(t, ws.Root())
	})

	t.Run("Default behavior without WithCleanup option", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		workspaceID := "test-default"

		// Create workspace without WithCleanup (should default to false)
		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
		)
		require.NoError(t, err)
		require.NoError(t, ws.Init(ctx))

		// Verify cleanup flag is false by default
		assert.False(t, ws.cleanupBeforeInit)
	})

	t.Run("WithCleanup(true) handles non-existent directory gracefully", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		workspaceID := "test-nonexistent"

		// Create workspace with cleanup when directory doesn't exist yet
		ws, err := New(v,
			WithID(workspaceID),
			WithTmpRoot(tmpDir),
			WithLogger(logger),
			WithCleanup(true),
		)
		require.NoError(t, err)

		// Init should succeed even though directory doesn't exist
		err = ws.Init(ctx)
		require.NoError(t, err)

		// Verify the directory was created
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

		// Create the directory and add some content
		rootDir := ws.rootDir()
		err = os.MkdirAll(rootDir, defaultDirPermissions)
		require.NoError(t, err)

		testFile := filepath.Join(rootDir, "test.txt")
		err = os.WriteFile(testFile, []byte("content"), 0o644)
		require.NoError(t, err)

		// Create subdirectory
		subDir := filepath.Join(rootDir, "subdir")
		err = os.MkdirAll(subDir, defaultDirPermissions)
		require.NoError(t, err)

		// Call cleanup method
		err = ws.cleanupExistingDir()
		require.NoError(t, err)

		// Verify directory is gone
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

		// Call cleanup method on non-existent directory
		err = ws.cleanupExistingDir()
		require.NoError(t, err)
	})
}
