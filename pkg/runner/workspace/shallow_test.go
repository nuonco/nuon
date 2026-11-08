package workspace_test

import (
	"context"
	"net/http/cgi"
	"net/http/httptest"
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

func TestWorkspaceShallowCloneOverHTTP(t *testing.T) {
	t.Setenv("IS_NUONCTL", "false")
	gitEnv := append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=Example", "-c", "user.email=example@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = dir
		cmd.Env = gitEnv
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", output)
		return strings.TrimSpace(string(output))
	}

	src := t.TempDir()
	git(src, "init", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(src, "payload"), []byte("first"), 0o600))
	git(src, "add", ".")
	git(src, "commit", "-m", "first")
	first := git(src, "rev-parse", "HEAD")
	git(src, "tag", "release-v1")
	require.NoError(t, os.WriteFile(filepath.Join(src, "payload"), []byte("second"), 0o600))
	git(src, "add", ".")
	git(src, "commit", "-m", "second")

	served := t.TempDir()
	git(served, "clone", "--bare", src, "sha-allowed.git")
	git(filepath.Join(served, "sha-allowed.git"), "config", "uploadpack.allowReachableSHA1InWant", "true")
	git(served, "clone", "--bare", src, "sha-denied.git")

	server := httptest.NewServer(&cgi.Handler{
		Path: filepath.Join(git(src, "--exec-path"), "git-http-backend"),
		Env:  append(gitEnv, "GIT_PROJECT_ROOT="+served, "GIT_HTTP_EXPORT_ALL=1"),
	})
	t.Cleanup(server.Close)

	newWorkspace := func(t *testing.T, fs afero.Fs, repo, ref string) workspace.Workspace {
		w, err := workspace.New(validator.New(), workspace.WithWorkspaceID("shallow"),
			workspace.WithTmpRoot(filepath.Join(t.TempDir(), "jobs")), workspace.WithFS(fs), workspace.WithLogger(zap.NewNop()),
			workspace.WithGitSource(&plantypes.GitSource{URL: server.URL + "/" + repo, Ref: ref, Path: "."}))
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, w.Cleanup(context.Background())) })
		return w
	}

	for _, backend := range []string{"disk", "memory"} {
		for _, tc := range []struct {
			name, repo, ref, contents string
			shallow                   bool
		}{
			{"branch", "sha-allowed.git", "main", "second", true},
			{"tag", "sha-allowed.git", "release-v1", "first", true},
			{"commit", "sha-allowed.git", first, "first", true},
			{"branch-without-sha-support", "sha-denied.git", "main", "second", true},
			{"commit-without-sha-support-falls-back", "sha-denied.git", first, "first", false},
		} {
			t.Run(backend+"/"+tc.name, func(t *testing.T) {
				var fs afero.Fs
				if backend == "memory" {
					fs = afero.NewMemMapFs()
				}
				w := newWorkspace(t, fs, tc.repo, tc.ref)
				require.NoError(t, w.Init(context.Background()))
				contents, err := afero.ReadFile(w.FS(), "payload")
				require.NoError(t, err)
				require.Equal(t, tc.contents, string(contents))
				require.Equal(t, tc.shallow, w.IsFile(".git/shallow"))
			})
		}

		t.Run(backend+"/missing-ref-does-not-fall-back", func(t *testing.T) {
			var fs afero.Fs = afero.NewOsFs()
			if backend == "memory" {
				fs = afero.NewMemMapFs()
			}
			w := newWorkspace(t, fs, "sha-allowed.git", "missing-branch")
			err := w.Init(context.Background())
			require.ErrorContains(t, err, "unable to clone repo")
			_, statErr := fs.Stat(w.Root())
			require.True(t, os.IsNotExist(statErr), "failed clone must remove its partial checkout")
		})
	}
}
