package workspace

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	runnerworkspace "github.com/nuonco/nuon/pkg/runner/workspace"
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
		contents, err := os.ReadFile(testFile)
		require.NoError(t, err)
		require.Equal(t, "test content", string(contents))
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

func TestClonedWorkspace(t *testing.T) {
	remote := t.TempDir()
	repo, err := git.PlainInitWithOptions(remote, &git.PlainInitOptions{InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("main")}})
	require.NoError(t, err)
	worktree, err := repo.Worktree()
	require.NoError(t, err)
	commit := func(contents string) plumbing.Hash {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(remote, "input.txt"), []byte(contents), 0o644))
		_, err := worktree.Add("input.txt")
		require.NoError(t, err)
		hash, err := worktree.Commit(contents, &git.CommitOptions{Author: &object.Signature{
			Name: "Acme", Email: "acme@example.com", When: time.Unix(1700000000, 0),
		}})
		require.NoError(t, err)
		return hash
	}
	base := commit("before\n")
	_, err = repo.CreateTag("v1", base, nil)
	require.NoError(t, err)
	head := commit("after\n")

	for _, memory := range []bool{false, true} {
		for ref, expected := range map[string]plumbing.Hash{"main": head, "v1": base, head.String(): head} {
			t.Run(fmt.Sprintf("memory=%v/ref=%s", memory, ref), func(t *testing.T) {
				var backing afero.Fs = afero.NewOsFs()
				if memory {
					backing = afero.NewMemMapFs()
				}
				parent := t.TempDir()
				ws, err := New(validator.New(), WithID("job"), WithTmpRoot(parent), WithFS(backing),
					WithLogger(zap.NewNop()), WithGitSource(&GitSource{URL: remote, Ref: ref}))
				require.NoError(t, err)
				require.NoError(t, ws.Init(context.Background()))
				t.Cleanup(func() { require.NoError(t, ws.Cleanup(context.Background())) })
				cloned, err := ws.Repository()
				require.NoError(t, err)
				actual, err := cloned.Head()
				require.NoError(t, err)
				require.Equal(t, expected, actual.Hash())
				contents, err := afero.ReadFile(ws.FS(), "input.txt")
				require.NoError(t, err)
				if expected == base {
					require.Equal(t, "before\n", string(contents))
				} else {
					require.Equal(t, "after\n", string(contents))
				}

				if !memory {
					neighbor := filepath.Join(parent, "neighbor")
					require.NoError(t, os.MkdirAll(filepath.Join(neighbor, ".git"), 0o700))
					require.NoError(t, os.WriteFile(filepath.Join(neighbor, ".git", "HEAD"), []byte(strings.Repeat("1", 40)+"\n"), 0o600))
					require.NoError(t, os.WriteFile(filepath.Join(neighbor, ".git", "config"), []byte("[remote \"origin\"]\nurl = canary\n"), 0o600))
					pinned := filepath.Join(parent, "pinned")
					require.NoError(t, os.Rename(ws.Root(), pinned))
					require.NoError(t, os.Symlink(neighbor, ws.Root()))
					require.NoError(t, os.Rename(filepath.Join(pinned, ".git"), filepath.Join(pinned, ".git-pinned")))
					require.NoError(t, os.Symlink(filepath.Join(neighbor, ".git"), filepath.Join(pinned, ".git")))
					actual, err = cloned.Head()
					require.NoError(t, err)
					require.Equal(t, expected, actual.Hash())
					cfg, err := cloned.Config()
					require.NoError(t, err)
					require.Equal(t, []string{remote}, cfg.Remotes["origin"].URLs)
					t.Cleanup(func() {
						canary, err := os.ReadFile(filepath.Join(neighbor, ".git", "HEAD"))
						require.NoError(t, err)
						require.Equal(t, strings.Repeat("1", 40)+"\n", string(canary))
					})
				}

				before, err := cloned.CommitObject(base)
				require.NoError(t, err)
				after, err := cloned.CommitObject(head)
				require.NoError(t, err)
				patch, err := before.Patch(after)
				require.NoError(t, err)
				require.Len(t, patch.Stats(), 1)
				require.Equal(t, "input.txt", patch.Stats()[0].Name)
				require.Equal(t, 1, patch.Stats()[0].Addition)
				require.Equal(t, 1, patch.Stats()[0].Deletion)
				require.Contains(t, patch.String(), "-before\n+after\n")
				require.NoError(t, ws.Close())
				require.NoError(t, ws.Close())
				_, err = ws.Repository()
				require.Error(t, err)
				_, err = backing.Stat(ws.Root())
				require.NoError(t, err)
				require.NoError(t, ws.Cleanup(context.Background()))
				_, err = backing.Stat(ws.Root())
				require.True(t, os.IsNotExist(err))
			})
		}
		for _, failure := range []struct{ name, url, ref string }{
			{"missing-repository", filepath.Join(t.TempDir(), "missing"), "main"},
			{"missing-ref", remote, "missing-branch"},
		} {
			t.Run(fmt.Sprintf("memory=%v/%s", memory, failure.name), func(t *testing.T) {
				var backing afero.Fs = afero.NewOsFs()
				if memory {
					backing = afero.NewMemMapFs()
				}
				parent := t.TempDir()
				ws, err := New(validator.New(), WithID("failed-clone"), WithTmpRoot(parent), WithFS(backing),
					WithLogger(zap.NewNop()), WithGitSource(&GitSource{URL: failure.url, Ref: failure.ref}))
				require.NoError(t, err)
				require.ErrorContains(t, ws.Init(context.Background()), "unable to clone repo")
				_, err = backing.Stat(ws.Root())
				require.True(t, os.IsNotExist(err), "failed clone must remove its partial checkout")
				require.NoError(t, ws.Close())
				require.NoError(t, ws.Cleanup(context.Background()))
			})
		}
	}
}

func TestAuthenticatedClones(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	gitBin, err := exec.LookPath("git")
	require.NoError(t, err)
	parent := t.TempDir()
	remote := filepath.Join(parent, "acme.git")
	repo, err := git.PlainInitWithOptions(remote, &git.PlainInitOptions{InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("main")}})
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)
	commit := func(contents string) plumbing.Hash {
		require.NoError(t, os.WriteFile(filepath.Join(remote, "payload"), []byte(contents), 0o600))
		_, err := wt.Add("payload")
		require.NoError(t, err)
		hash, err := wt.Commit("fixture", &git.CommitOptions{Author: &object.Signature{Name: "Acme", Email: "acme@example.com", When: time.Unix(1700000000, 0)}})
		require.NoError(t, err)
		return hash
	}
	_, err = repo.CreateTag("v1", commit("tag contents"), nil)
	require.NoError(t, err)
	require.NoError(t, repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), commit("feature contents"))))
	hash := commit("hash contents")
	backend := &cgi.Handler{Path: gitBin, Args: []string{"http-backend"}, Env: []string{"GIT_PROJECT_ROOT=" + parent, "GIT_HTTP_EXPORT_ALL=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "acme" || password != "DISPOSABLE_CLONE_TOKEN" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL + "/acme.git")
	require.NoError(t, err)
	u.User = url.UserPassword("acme", "DISPOSABLE_CLONE_TOKEN")
	for _, runner := range []bool{false, true} {
		for ref, expected := range map[string]string{"feature": "feature contents", "v1": "tag contents", hash.String(): "hash contents"} {
			t.Run(fmt.Sprintf("runner=%v/ref=%s", runner, ref), func(t *testing.T) {
				core, logs := observer.New(zap.DebugLevel)
				var fs afero.Fs
				if runner {
					ws, err := runnerworkspace.New(validator.New(), runnerworkspace.WithWorkspaceID("private"), runnerworkspace.WithTmpRoot(t.TempDir()), runnerworkspace.WithLogger(zap.New(core)), runnerworkspace.WithGitSource(&plantypes.GitSource{URL: u.String(), Ref: ref, Path: "."}))
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, ws.Cleanup(context.Background())) })
					require.NoError(t, ws.Init(context.Background()))
					fs = ws.FS()
				} else {
					ws, err := New(validator.New(), WithID("private"), WithTmpRoot(t.TempDir()), WithLogger(zap.New(core)), WithGitSource(&GitSource{URL: u.String(), Ref: ref}))
					require.NoError(t, err)
					t.Cleanup(func() { require.NoError(t, ws.Cleanup(context.Background())) })
					require.NoError(t, ws.Init(context.Background()))
					fs = ws.FS()
				}
				contents, err := afero.ReadFile(fs, "payload")
				require.NoError(t, err)
				require.Equal(t, expected, string(contents))
				metadata, err := afero.ReadFile(fs, ".git/config")
				require.NoError(t, err)
				require.NotContains(t, string(metadata), "DISPOSABLE_CLONE_TOKEN")
				require.Contains(t, string(metadata), server.URL+"/acme.git")
				for _, entry := range logs.All() {
					require.NotContains(t, fmt.Sprint(entry.Message, entry.ContextMap()), "DISPOSABLE_CLONE_TOKEN")
				}
			})
		}
	}
}

func TestWorkspaceRejectsTraversal(t *testing.T) {
	for _, name := range []string{"..", "../neighbor", "/tmp/job", "nested/job", `nested\job`} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			ws, err := New(validator.New(), WithID(name), WithTmpRoot(parent), WithLogger(zap.NewNop()))
			require.NoError(t, err)
			require.Error(t, ws.Init(context.Background()))
			entries, err := os.ReadDir(parent)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
	for _, path := range []string{"../../var/run/secrets/kubernetes.io/serviceaccount", "component/../../outside", "/../../outside"} {
		t.Run(path, func(t *testing.T) {
			parent := t.TempDir()
			ws, err := New(validator.New(), WithID("build"), WithTmpRoot(parent), WithLogger(zap.NewNop()), WithGitSource(&GitSource{Path: path}))
			require.NoError(t, err)
			require.ErrorContains(t, ws.Init(context.Background()), "invalid repository path")
			entries, err := os.ReadDir(parent)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}
