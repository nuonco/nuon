package helpers

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	object := func(commit, root string) string {
		return git(src, "rev-parse", commit+":"+root)
	}
	gitInput := func(input string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = src
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}
	bomb := gitInput("100644 blob "+object(inside, "app/README.md")+"\tf\n", "mktree")
	for range 10 {
		var entries strings.Builder
		for i := range 10 {
			fmt.Fprintf(&entries, "040000 tree %s\td%d\n", bomb, i)
		}
		bomb = gitInput(entries.String(), "mktree")
	}
	link := gitInput("../other/main.tf", "hash-object", "-w", "--stdin")
	tools := gitInput("120000 blob "+link+"\tmain.tf\n", "mktree")
	readme, err := hex.DecodeString(object(inside, "app/README.md"))
	require.NoError(t, err)
	legacy := gitInput("100664 README.md\x00"+string(readme), "hash-object", "-t", "tree", "--literally", "-w", "--stdin")
	rootEntries := git(src, "ls-tree", inside) + "\n040000 tree " + bomb + "\tbomb\n040000 tree " + tools + "\ttools\n040000 tree " + legacy + "\tlegacy\n"
	shapes := gitInput("shapes", "commit-tree", gitInput(rootEntries, "mktree"), "-p", inside)
	git(src, "update-ref", "refs/heads/shapes", shapes)

	plain := filepath.Join(dir, "plain.git")
	filtered := filepath.Join(dir, "filtered.git")
	git(dir, "clone", "-q", "--bare", src, plain)
	git(dir, "clone", "-q", "--bare", src, filtered)
	git(filtered, "config", "uploadpack.allowFilter", "true")
	git(filtered, "config", "uploadpack.allowAnySHA1InWant", "true")

	tests := map[string]struct {
		server          string
		commit          string
		root            string
		wantScope       string
		wantFetchErr    error
		wantScopeErr    error
		wantUnsupported string
		wantNoBlobs     bool
		serverCaps      *capability.List
		maxBytes        int64
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
			server:          filtered,
			commit:          first,
			root:            "",
			wantScope:       "git:040000:" + object(first, ""),
			wantUnsupported: UnsupportedSubmodule,
		},
		"slash means the repository root": {
			server:          filtered,
			commit:          first,
			root:            "/",
			wantScope:       "git:040000:" + object(first, ""),
			wantUnsupported: UnsupportedSubmodule,
		},
		"leading slash is relative to the repository": {
			server:    filtered,
			commit:    inside,
			root:      "/app/chart",
			wantScope: "git:040000:" + object(inside, "app/chart"),
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
		"symlink inside the directory": {
			server:    filtered,
			commit:    shapes,
			root:      "tools",
			wantScope: "git:040000:" + object(shapes, "tools"),
		},
		"repeated subtrees are walked once": {
			server:    filtered,
			commit:    shapes,
			root:      "bomb",
			wantScope: "git:040000:" + object(shapes, "bomb"),
		},
		"non-canonical tree under the directory": {
			server:          filtered,
			commit:          shapes,
			root:            "legacy",
			wantScope:       "git:040000:" + object(shapes, "legacy"),
			wantUnsupported: UnsupportedNonCanonical,
		},
		"submodule under the directory": {
			server:          filtered,
			commit:          first,
			root:            "vendor",
			wantScope:       "git:040000:" + object(first, "vendor"),
			wantUnsupported: UnsupportedSubmodule,
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
		"pack over the size limit": {
			server:       filtered,
			commit:       inside,
			maxBytes:     64,
			wantFetchErr: ErrSourceTreeTooLarge,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			maxBytes := int64(maxSourceTreeBytes)
			if tc.maxBytes > 0 {
				maxBytes = tc.maxBytes
			}
			tree, err := fetchSourceTree(context.Background(), "file://"+tc.server, nil, tc.commit, tc.serverCaps, maxBytes)
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

			unsupported, err := tree.Unsupported(context.Background(), tc.root)
			require.NoError(t, err)
			assert.Equal(t, tc.wantUnsupported, unsupported)

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

	t.Run("pack size counts inflated objects, delta payloads, and delta targets", func(t *testing.T) {
		cmd := exec.Command("git", "pack-objects", "--revs", "--stdout", "--filter=blob:none")
		cmd.Dir = src
		cmd.Stdin = strings.NewReader(inside + "\n")
		pack, err := cmd.Output()
		require.NoError(t, err)
		objects := strings.Fields(git(src, "rev-list", "--objects", "--no-object-names", "--filter=blob:none", inside))
		var total int64
		for _, size := range strings.Fields(gitInput(strings.Join(objects, "\n")+"\n", "cat-file", "--batch-check=%(objectsize)")) {
			n, err := strconv.ParseInt(size, 10, 64)
			require.NoError(t, err)
			total += n
		}
		packPath := filepath.Join(t.TempDir(), "objects.pack")
		require.NoError(t, os.WriteFile(packPath, pack, 0o600))
		git(src, "index-pack", packPath)
		deltas := 0
		for _, line := range strings.Split(git(src, "verify-pack", "-v", strings.TrimSuffix(packPath, ".pack")+".idx"), "\n") {
			if fields := strings.Fields(line); len(fields) == 7 {
				n, err := strconv.ParseInt(fields[2], 10, 64)
				require.NoError(t, err)
				total += n
				deltas++
			}
		}
		require.NotZero(t, deltas)
		require.NoError(t, checkPackSize(bytes.NewReader(pack), total))
		require.ErrorIs(t, checkPackSize(bytes.NewReader(pack), total-1), ErrSourceTreeTooLarge)
	})
}
