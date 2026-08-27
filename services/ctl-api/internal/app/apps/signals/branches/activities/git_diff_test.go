package activities

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

func TestDiffCommitsChangedPaths(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644))
	run("add", "a.txt")
	run("commit", "-m", "base")

	repo, err := git.PlainOpen(dir)
	require.NoError(t, err)
	baseRef, err := repo.Head()
	require.NoError(t, err)
	baseSHA := baseRef.Hash().String()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("two\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new\n"), 0o644))
	run("add", "a.txt", "b.txt")
	run("commit", "-m", "head")

	headRef, err := repo.Head()
	require.NoError(t, err)
	headSHA := headRef.Hash().String()

	result, err := diffCommits(context.Background(), repo, baseSHA, headSHA)
	require.NoError(t, err)
	require.Equal(t, baseSHA, result.BaseSHA)
	require.Equal(t, headSHA, result.HeadSHA)
	require.Equal(t, 2, result.FilesChanged)
	require.ElementsMatch(t, []string{"a.txt", "b.txt"}, result.ChangedPaths)
	require.Contains(t, result.Patch, "a.txt")
	require.Contains(t, result.Patch, "b.txt")
}
