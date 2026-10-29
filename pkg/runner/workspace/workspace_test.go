package workspace_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

func TestWorkspaceCheckoutOnDiskAndMemory(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	repo := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=Example", "-c", "user.email=example@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
		return strings.TrimSpace(string(output))
	}
	git("init", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "payload"), []byte("first"), 0o600))
	git("add", ".")
	git("commit", "-m", "first")
	first := git("rev-parse", "HEAD")
	git("tag", "release-v1")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "payload"), []byte("second"), 0o600))
	git("add", ".")
	git("commit", "-m", "second")

	for _, backend := range []string{"disk", "memory"} {
		for _, selection := range []struct{ ref, contents string }{
			{"main", "second"}, {"release-v1", "first"}, {first, "first"},
		} {
			t.Run(backend+"/"+selection.ref, func(t *testing.T) {
				root := filepath.Join(t.TempDir(), "jobs")
				var fs afero.Fs
				if backend == "memory" {
					fs = afero.NewMemMapFs()
				}
				w, err := workspace.New(validator.New(), workspace.WithWorkspaceID("checkout"),
					workspace.WithTmpRoot(root), workspace.WithFS(fs), workspace.WithLogger(zap.NewNop()),
					workspace.WithGitSource(&plantypes.GitSource{URL: repo, Ref: selection.ref, Path: "."}))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, w.Cleanup(context.Background())) })
				require.NoError(t, w.Init(context.Background()))
				contents, err := afero.ReadFile(w.FS(), "payload")
				require.NoError(t, err)
				require.Equal(t, selection.contents, string(contents))
				require.True(t, w.IsFile("payload"))
				require.True(t, w.IsDir(".git"))
				if backend == "disk" {
					require.NoError(t, w.FS().(afero.Linker).SymlinkIfPossible("payload", "safe-link"))
					linked, err := afero.ReadFile(w.FS(), "safe-link")
					require.NoError(t, err)
					require.Equal(t, selection.contents, string(linked))
				} else {
					_, err := os.Stat(root)
					require.True(t, os.IsNotExist(err), "memory clone must not create a host workspace")
				}
				require.NoError(t, w.Cleanup(context.Background()))
				require.Nil(t, w.FS())
				if fs != nil {
					_, err := fs.Stat(filepath.Join(root, "workspace-checkout"))
					require.True(t, os.IsNotExist(err))
				}
			})
		}
		for _, failure := range []struct{ name, url, ref string }{
			{"missing-repository", filepath.Join(t.TempDir(), "missing"), "main"},
			{"missing-ref", repo, "missing-branch"},
		} {
			t.Run(backend+"/"+failure.name, func(t *testing.T) {
				var fs afero.Fs = afero.NewOsFs()
				if backend == "memory" {
					fs = afero.NewMemMapFs()
				}
				root := t.TempDir()
				w, err := workspace.New(validator.New(), workspace.WithWorkspaceID("failed-clone"),
					workspace.WithTmpRoot(root), workspace.WithFS(fs), workspace.WithLogger(zap.NewNop()),
					workspace.WithGitSource(&plantypes.GitSource{URL: failure.url, Ref: failure.ref, Path: "."}))
				require.NoError(t, err)
				require.ErrorContains(t, w.Init(context.Background()), "unable to clone repo")
				_, err = fs.Stat(w.Root())
				require.True(t, os.IsNotExist(err), "failed clone must remove its partial checkout")
				require.NoError(t, w.Cleanup(context.Background()))
			})
		}
	}
}

func TestWorkspaceRetainsPinnedRoot(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	root := t.TempDir()
	w, err := workspace.New(validator.New(), workspace.WithWorkspaceID("pinned"),
		workspace.WithTmpRoot(root), workspace.WithLogger(zap.NewNop()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, w.Cleanup(context.Background())) })
	require.NoError(t, w.Init(context.Background()))
	require.NoError(t, afero.WriteFile(w.FS(), "payload", []byte("owned"), 0o600))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "payload"), []byte("canary"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "outside-only"), []byte("secret"), 0o600))
	require.NoError(t, os.Rename(w.Root(), filepath.Join(root, "moved")))
	require.NoError(t, os.Symlink(outside, w.Root()))
	contents, err := afero.ReadFile(w.FS(), "payload")
	require.NoError(t, err)
	require.Equal(t, "owned", string(contents))
	require.False(t, w.IsFile("outside-only"))
	require.False(t, w.IsFile("../outside-only"))
	require.NoError(t, w.Cleanup(context.Background()))
	contents, err = os.ReadFile(filepath.Join(outside, "payload"))
	require.NoError(t, err)
	require.Equal(t, "canary", string(contents))
	_, err = os.Stat(filepath.Join(outside, "outside-only"))
	require.NoError(t, err)
}
