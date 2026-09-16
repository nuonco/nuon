package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIsFullCommitSHA(t *testing.T) {
	for _, tc := range []struct {
		name string
		ref  string
		want bool
	}{
		{"full sha", "3a5f1b2c4d6e8f0a1b2c3d4e5f60718293a4b5c6", true},
		{"short sha", "3a5f1b2", false},
		{"branch that looks like hex", "deadbeef", false},
		{"branch containing hex", "feature/deadbeef", false},
		{"uppercase sha", "3A5F1B2C4D6E8F0A1B2C3D4E5F60718293A4B5C6", false},
		{"empty", "", false},
		{"main", "main", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsFullCommitSHA(tc.ref))
		})
	}
}

// seedRepo builds a repo with two commits and returns its path, the branch
// name, and both commit hashes in order.
func seedRepo(t *testing.T) (string, string, plumbing.Hash, plumbing.Hash) {
	t.Helper()

	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	commit := func(name, contents string) plumbing.Hash {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644))
		_, err := wt.Add(name)
		require.NoError(t, err)

		h, err := wt.Commit("add "+name, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@nuon.co",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
		return h
	}

	first := commit("first.txt", "first")
	second := commit("second.txt", "second")

	head, err := repo.Head()
	require.NoError(t, err)

	return dir, head.Name().Short(), first, second
}

func countCommits(t *testing.T, dir string) int {
	t.Helper()

	repo, err := git.PlainOpen(dir)
	require.NoError(t, err)

	iter, err := repo.Log(&git.LogOptions{})
	require.NoError(t, err)
	defer iter.Close()

	count := 0
	require.NoError(t, iter.ForEach(func(*object.Commit) error {
		count++
		return nil
	}))

	return count
}

// isShallow reports whether the clone recorded a shallow boundary, which is
// how git marks a depth-limited fetch.
func isShallow(t *testing.T, dir string) bool {
	t.Helper()

	_, err := os.Stat(filepath.Join(dir, ".git", "shallow"))
	return err == nil
}

func cloneInto(t *testing.T, srcDir, ref string) *Workspace {
	t.Helper()

	ws, err := New(validator.New(),
		WithID("clone-test"),
		WithTmpRoot(t.TempDir()),
		WithLogger(zap.NewNop()),
		WithGitSource(&GitSource{URL: srcDir, Ref: ref}),
	)
	require.NoError(t, err)
	require.NoError(t, ws.Init(context.Background()))

	return ws
}

func TestCloneRefRouting(t *testing.T) {
	srcDir, branch, first, second := seedRepo(t)

	t.Run("branch ref checks out the branch tip, shallow", func(t *testing.T) {
		ws := cloneInto(t, srcDir, branch)

		repo, err := git.PlainOpen(ws.Root())
		require.NoError(t, err)
		head, err := repo.Head()
		require.NoError(t, err)

		assert.Equal(t, second, head.Hash())
		assert.FileExists(t, filepath.Join(ws.Root(), "second.txt"))
		assert.True(t, isShallow(t, ws.Root()), "expected a depth-limited clone")
	})

	t.Run("full sha checks out that exact commit", func(t *testing.T) {
		ws := cloneInto(t, srcDir, first.String())

		repo, err := git.PlainOpen(ws.Root())
		require.NoError(t, err)
		head, err := repo.Head()
		require.NoError(t, err)

		assert.Equal(t, first, head.Hash())
		assert.FileExists(t, filepath.Join(ws.Root(), "first.txt"))
		assert.NoFileExists(t, filepath.Join(ws.Root(), "second.txt"))
	})

	t.Run("full sha keeps history rather than going shallow", func(t *testing.T) {
		ws := cloneInto(t, srcDir, second.String())

		assert.False(t, isShallow(t, ws.Root()))
		assert.Equal(t, 2, countCommits(t, ws.Root()))
	})
}
