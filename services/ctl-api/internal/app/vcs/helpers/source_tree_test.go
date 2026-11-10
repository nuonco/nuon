package helpers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchSourceTree(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	git := func(repo string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}
	write := func(name, contents string) {
		t.Helper()
		p := filepath.Join(src, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(contents), 0o644))
	}

	require.NoError(t, os.Mkdir(src, 0o755))
	git(src, "init", "-q")
	write("app/chart/values.yaml", "replicas: 1\n")
	write("app/README.md", "app\n")
	write("other/main.tf", "# other\n")
	require.NoError(t, os.Symlink("app", filepath.Join(src, "linked")))
	git(src, "add", "-A")
	git(src, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,vendor/lib")
	git(src, "commit", "-qm", "first")
	first := git(src, "rev-parse", "HEAD")
	write("other/main.tf", "# other changed\n")
	git(src, "commit", "-qam", "outside app")
	outside := git(src, "rev-parse", "HEAD")
	write("app/chart/values.yaml", "replicas: 2\n")
	git(src, "commit", "-qam", "inside app")
	inside := git(src, "rev-parse", "HEAD")

	plain := filepath.Join(dir, "plain.git")
	filtered := filepath.Join(dir, "filtered.git")
	git(dir, "clone", "-q", "--bare", src, plain)
	git(dir, "clone", "-q", "--bare", src, filtered)
	git(filtered, "config", "uploadpack.allowFilter", "true")
	git(filtered, "config", "uploadpack.allowAnySHA1InWant", "true")

	object := func(commit, root string) string {
		return git(src, "rev-parse", commit+":"+root)
	}

	tests := map[string]struct {
		server        string
		commit        string
		root          string
		wantScope     string
		wantFetchErr  error
		wantScopeErr  error
		wantSubmodule bool
		wantNoBlobs   bool
		serverCaps    *capability.List
	}{
		"filtering server fetches an older commit without blobs": {
			server:      filtered,
			commit:      first,
			root:        "app",
			wantScope:   "git:040000:" + object(first, "app"),
			wantNoBlobs: true,
		},
		"assumed GitHub capabilities fetch without blobs": {
			server:      filtered,
			commit:      first,
			root:        "app",
			wantScope:   "git:040000:" + object(first, "app"),
			wantNoBlobs: true,
			serverCaps:  githubUploadPackCapabilities(),
		},
		"directory keeps its tree when only other paths change": {
			server:    filtered,
			commit:    outside,
			root:      "app",
			wantScope: "git:040000:" + object(first, "app"),
		},
		"directory gets a new tree when a file under it changes": {
			server:    filtered,
			commit:    inside,
			root:      "app/chart",
			wantScope: "git:040000:" + object(inside, "app/chart"),
		},
		"repository root": {
			server:        filtered,
			commit:        first,
			root:          "",
			wantScope:     "git:040000:" + object(first, ""),
			wantSubmodule: true,
		},
		"file root": {
			server:    filtered,
			commit:    first,
			root:      "app/chart/values.yaml",
			wantScope: "git:100644:" + object(first, "app/chart/values.yaml"),
		},
		"path below a symlink pins the symlink": {
			server:    filtered,
			commit:    first,
			root:      "linked/chart",
			wantScope: "git:120000:" + object(first, "linked"),
		},
		"submodule under the directory": {
			server:        filtered,
			commit:        first,
			root:          "vendor",
			wantScope:     "git:040000:" + object(first, "vendor"),
			wantSubmodule: true,
		},
		"missing directory": {
			server:       filtered,
			commit:       first,
			root:         "missing",
			wantScopeErr: ErrSourceDirectoryMissing,
		},
		"unknown commit": {
			server:       filtered,
			commit:       "2222222222222222222222222222222222222222",
			wantFetchErr: ErrSourceCommitNotFound,
		},
		"server without filtering still serves the branch tip": {
			server:    plain,
			commit:    inside,
			root:      "app/chart",
			wantScope: "git:040000:" + object(inside, "app/chart"),
		},
		"server without filtering refuses an older commit": {
			server:       plain,
			commit:       first,
			wantFetchErr: ErrSourceCommitNotFound,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tree, err := fetchSourceTree(context.Background(), "file://"+tc.server, nil, tc.commit, tc.serverCaps)
			if tc.wantFetchErr != nil {
				require.ErrorIs(t, err, tc.wantFetchErr)
				return
			}
			require.NoError(t, err)

			scope, err := tree.Scope(tc.root)
			if tc.wantScopeErr != nil {
				require.ErrorIs(t, err, tc.wantScopeErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantScope, scope)

			submodule, err := tree.HasSubmodule(tc.root)
			require.NoError(t, err)
			assert.Equal(t, tc.wantSubmodule, submodule)

			if tc.wantNoBlobs {
				blobs, err := tree.s.IterEncodedObjects(plumbing.BlobObject)
				require.NoError(t, err)
				count := 0
				require.NoError(t, blobs.ForEach(func(plumbing.EncodedObject) error {
					count++
					return nil
				}))
				assert.Zero(t, count)
			}
		})
	}
}
