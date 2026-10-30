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
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/rootfs"
	runnerworkspace "github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/pkg/workspace"
)

type job interface {
	Init(context.Context) error
	Cleanup(context.Context) error
	Root() string
	FS() afero.Fs
	IsDir(string) bool
	IsFile(string) bool
}

type jobOptions struct {
	root, id, url, ref, path string
	fs                       afero.Fs
	logger                   *zap.Logger
}

type jobFactory struct {
	name string
	new  func(*testing.T, jobOptions) job
}

var factories = []jobFactory{
	{"shared", func(t *testing.T, opts jobOptions) job {
		var src *workspace.GitSource
		if opts.url != "" {
			src = &workspace.GitSource{URL: opts.url, Ref: opts.ref, Path: opts.path}
		}
		w, err := workspace.New(validator.New(), workspace.WithID(opts.id), workspace.WithTmpRoot(opts.root),
			workspace.WithGitSource(src), workspace.WithLogger(opts.logger), workspace.WithFS(opts.fs))
		require.NoError(t, err)
		return w
	}},
	{"runner", func(t *testing.T, opts jobOptions) job {
		var src *plantypes.GitSource
		if opts.url != "" {
			if opts.path == "" {
				opts.path = "."
			}
			src = &plantypes.GitSource{URL: opts.url, Ref: opts.ref, Path: opts.path}
		}
		w, err := runnerworkspace.New(validator.New(), runnerworkspace.WithWorkspaceID(opts.id), runnerworkspace.WithTmpRoot(opts.root),
			runnerworkspace.WithGitSource(src), runnerworkspace.WithLogger(opts.logger), runnerworkspace.WithFS(opts.fs))
		require.NoError(t, err)
		return w
	}},
}

func gitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=Example", "-c", "user.email=example@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
	return strings.TrimSpace(string(output))
}

func fixture(t *testing.T) (dir, first, second string) {
	t.Helper()
	dir = t.TempDir()
	gitCommand(t, dir, "init", "-b", "main")
	require.NoError(t, os.Mkdir(filepath.Join(dir, "dir"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dir/payload"), []byte("first"), 0o600))
	require.NoError(t, os.Symlink("payload", filepath.Join(dir, "dir/safe-link")))
	gitCommand(t, dir, "add", ".")
	gitCommand(t, dir, "commit", "-m", "first")
	first = gitCommand(t, dir, "rev-parse", "HEAD")
	gitCommand(t, dir, "tag", "release-v1")
	gitCommand(t, dir, "branch", "selected")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dir/payload"), []byte("second"), 0o600))
	gitCommand(t, dir, "add", ".")
	gitCommand(t, dir, "commit", "-m", "second")
	second = gitCommand(t, dir, "rev-parse", "HEAD")
	return dir, first, second
}

func TestPublicInitSelectsCheckoutAndKeepsSafeLinks(t *testing.T) {
	dir, first, second := fixture(t)
	for _, factory := range factories {
		for _, tc := range []struct{ ref, content string }{{first, "first"}, {second, "second"}, {"selected", "first"}, {"release-v1", "first"}} {
			t.Run(factory.name+"/"+tc.ref, func(t *testing.T) {
				w := factory.new(t, jobOptions{root: t.TempDir(), id: "job", url: dir, ref: tc.ref, path: "/dir", logger: zap.NewNop()})
				require.NoError(t, w.Init(t.Context()))
				t.Cleanup(func() { require.NoError(t, w.Cleanup(context.Background())) })
				require.IsType(t, &rootfs.FS{}, w.FS())
				require.True(t, w.IsDir("/dir"))
				require.True(t, w.IsFile("/dir/payload"))
				require.False(t, w.IsDir(".."))
				data, err := afero.ReadFile(w.FS(), "dir/safe-link")
				require.NoError(t, err)
				require.Equal(t, tc.content, string(data))
				target, err := w.FS().(afero.LinkReader).ReadlinkIfPossible("dir/safe-link")
				require.NoError(t, err)
				require.Equal(t, "payload", target)
				info, err := os.Stat(w.Root())
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
				switch w := w.(type) {
				case *workspace.Workspace:
					require.Equal(t, filepath.Join(w.Root(), "dir"), w.SourceDir())
				case interface {
					Root() string
					Source() *runnerworkspace.Source
				}:
					require.Equal(t, filepath.Join(w.Root(), "dir"), w.Source().AbsPath())
				}
			})
		}
	}
}

func TestPublicInitRejectsTraversalAndSymlinkRoots(t *testing.T) {
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			for _, id := range []string{"..", "../neighbor", "/absolute", "job/../../neighbor", `job\neighbor`} {
				w := factory.new(t, jobOptions{root: t.TempDir(), id: id, logger: zap.NewNop()})
				require.Error(t, w.Init(t.Context()), id)
				require.Nil(t, w.FS())
			}
			for _, path := range []string{"../neighbor", "/../neighbor", "dir/../../neighbor"} {
				w := factory.new(t, jobOptions{root: t.TempDir(), id: "job", url: "unused", ref: "main", path: path, logger: zap.NewNop()})
				require.Error(t, w.Init(t.Context()), path)
				require.Nil(t, w.FS())
			}
			root, neighbor := t.TempDir(), t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "canary"), []byte("neighbor"), 0o600))
			w := factory.new(t, jobOptions{root: root, id: "job", fs: afero.NewOsFs(), logger: zap.NewNop()})
			require.NoError(t, os.Symlink(neighbor, w.Root()))
			require.Error(t, w.Init(t.Context()))
			require.Nil(t, w.FS())
			data, err := os.ReadFile(filepath.Join(neighbor, "canary"))
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(data))
		})
	}
}

func TestPublicInitPinnedAcrossRootReplacement(t *testing.T) {
	dir, first, _ := fixture(t)
	for _, factory := range factories {
		for _, stage := range []string{"cloning repository", "fetching working tree"} {
			t.Run(factory.name+"/"+stage, func(t *testing.T) {
				root, neighbor := t.TempDir(), t.TempDir()
				require.NoError(t, os.Mkdir(filepath.Join(neighbor, "dir"), 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(neighbor, "dir/payload"), []byte("neighbor"), 0o600))
				var w job
				swapped := false
				core, _ := observer.New(zapcore.DebugLevel)
				logger := zap.New(core, zap.Hooks(func(entry zapcore.Entry) error {
					if entry.Message == stage && !swapped {
						swapped = true
						require.NoError(t, os.Rename(w.Root(), w.Root()+"-saved"))
						require.NoError(t, os.Symlink(neighbor, w.Root()))
					}
					return nil
				}))
				w = factory.new(t, jobOptions{root: root, id: "job", url: dir, ref: first, fs: afero.NewOsFs(), logger: logger})
				require.NoError(t, w.Init(t.Context()))
				require.True(t, swapped)
				borrowed := w.FS()
				data, err := afero.ReadFile(borrowed, "dir/payload")
				require.NoError(t, err)
				require.Equal(t, "first", string(data))
				require.FileExists(t, filepath.Join(w.Root()+"-saved", ".git/HEAD"))
				data, err = os.ReadFile(filepath.Join(neighbor, "dir/payload"))
				require.NoError(t, err)
				require.Equal(t, "neighbor", string(data))
				require.NoError(t, w.Cleanup(t.Context()))
				_, err = borrowed.Open("dir/payload")
				require.Error(t, err, "Cleanup must close borrowed job handles")
				require.FileExists(t, filepath.Join(neighbor, "dir/payload"))
			})
		}
	}
}

func TestPublicInitPreservesUnsafeLinkEvidenceWithoutFollowingIt(t *testing.T) {
	dir, _, _ := fixture(t)
	neighbor := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(neighbor, "canary"), []byte("neighbor"), 0o600))
	for name, target := range map[string]string{"absolute": filepath.Join(neighbor, "canary"), "relative": "../../canary"} {
		require.NoError(t, os.Symlink(target, filepath.Join(dir, name)))
	}
	gitCommand(t, dir, "add", ".")
	gitCommand(t, dir, "commit", "-m", "links")
	commit := gitCommand(t, dir, "rev-parse", "HEAD")
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			w := factory.new(t, jobOptions{root: t.TempDir(), id: "job", url: dir, ref: commit, logger: zap.NewNop()})
			require.NoError(t, w.Init(t.Context()))
			defer w.Cleanup(context.Background())
			for _, name := range []string{"absolute", "relative"} {
				target, err := w.FS().(afero.LinkReader).ReadlinkIfPossible(name)
				require.NoError(t, err)
				expected, err := os.Readlink(filepath.Join(dir, name))
				require.NoError(t, err)
				require.Equal(t, expected, target)
				_, err = afero.ReadFile(w.FS(), name)
				require.Error(t, err)
				require.Error(t, afero.WriteFile(w.FS(), name, []byte("bad"), 0o600))
			}
			data, err := os.ReadFile(filepath.Join(neighbor, "canary"))
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(data))
		})
	}
}

func TestPublicInitSymlinkedCheckoutDirectoryDoesNotOverwriteNeighbor(t *testing.T) {
	dir, first, _ := fixture(t)
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			root := t.TempDir()
			neighbor := filepath.Join(root, "neighbor")
			require.NoError(t, os.Mkdir(neighbor, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "payload"), []byte("neighbor"), 0o600))
			var w job
			injected := false
			core, _ := observer.New(zapcore.DebugLevel)
			logger := zap.New(core, zap.Hooks(func(entry zapcore.Entry) error {
				if entry.Message == "fetching working tree" && !injected {
					injected = true
					require.NoError(t, os.Symlink("../neighbor", filepath.Join(w.Root(), "dir")))
				}
				return nil
			}))
			w = factory.new(t, jobOptions{root: root, id: "job", url: dir, ref: first, logger: logger})
			require.NoError(t, w.Init(t.Context()))
			defer w.Cleanup(context.Background())
			require.True(t, injected)
			info, err := os.Lstat(filepath.Join(w.Root(), "dir"))
			require.NoError(t, err)
			require.True(t, info.IsDir())
			data, err := afero.ReadFile(w.FS(), "dir/payload")
			require.NoError(t, err)
			require.Equal(t, "first", string(data))
			data, err = os.ReadFile(filepath.Join(neighbor, "payload"))
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(data))
		})
	}
}

func TestPublicInitGitStoragePinnedAcrossDotgitReplacement(t *testing.T) {
	dir, first, _ := fixture(t)
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			root, neighbor := t.TempDir(), t.TempDir()
			saved := filepath.Join(root, "dotgit-saved")
			require.NoError(t, os.WriteFile(filepath.Join(neighbor, "HEAD"), []byte("neighbor"), 0o600))
			var w job
			swapped := false
			core, _ := observer.New(zapcore.DebugLevel)
			logger := zap.New(core, zap.Hooks(func(entry zapcore.Entry) error {
				if entry.Message == "fetching working tree" && !swapped {
					swapped = true
					require.NoError(t, os.Rename(filepath.Join(w.Root(), ".git"), saved))
					require.NoError(t, os.Symlink(neighbor, filepath.Join(w.Root(), ".git")))
				}
				return nil
			}))
			w = factory.new(t, jobOptions{root: root, id: "job", url: dir, ref: first, logger: logger})
			require.NoError(t, w.Init(t.Context()))
			require.True(t, swapped)
			defer w.Cleanup(context.Background())
			data, err := os.ReadFile(filepath.Join(saved, "HEAD"))
			require.NoError(t, err)
			require.Equal(t, first, strings.TrimSpace(string(data)))
			data, err = os.ReadFile(filepath.Join(neighbor, "HEAD"))
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(data))
		})
	}
}

func TestPublicInitMemoryNamespaceAndFailureLifecycle(t *testing.T) {
	t.Setenv("IS_NUONCTL", "")
	dir, _, _ := fixture(t)
	gitCommand(t, dir, "rm", "dir/safe-link")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dir/payload"), []byte("first"), 0o600))
	gitCommand(t, dir, "add", ".")
	gitCommand(t, dir, "commit", "-m", "plain memory fixture")
	first := gitCommand(t, dir, "rev-parse", "HEAD")
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			memory := afero.NewMemMapFs()
			root := filepath.Join(t.TempDir(), "memory-only")
			neighbor := filepath.Join(root, "neighbor", "canary")
			require.NoError(t, memory.MkdirAll(filepath.Dir(neighbor), 0o700))
			require.NoError(t, afero.WriteFile(memory, neighbor, []byte("neighbor"), 0o600))
			w := factory.new(t, jobOptions{root: root, id: "job", fs: memory, url: dir, ref: first, logger: zap.NewNop()})
			require.NoError(t, w.Init(t.Context()))
			data, err := afero.ReadFile(w.FS(), "dir/payload")
			require.NoError(t, err)
			require.Equal(t, "first", string(data))
			data, err = afero.ReadFile(memory, filepath.Join(w.Root(), "dir/payload"))
			require.NoError(t, err)
			require.Equal(t, "first", string(data))
			require.NoDirExists(t, root)
			require.NoError(t, w.Cleanup(t.Context()))
			data, err = afero.ReadFile(memory, neighbor)
			require.NoError(t, err)
			require.Equal(t, "neighbor", string(data))
			var other job
			var borrowed afero.Fs
			core, _ := observer.New(zapcore.DebugLevel)
			logger := zap.New(core, zap.Hooks(func(entry zapcore.Entry) error {
				if entry.Message == "cloning repository" {
					borrowed = other.FS()
				}
				return nil
			}))
			other = factory.new(t, jobOptions{root: root, id: "failed", url: dir, ref: strings.Repeat("f", 40), logger: logger})
			require.Error(t, other.Init(t.Context()))
			require.Nil(t, other.FS())
			require.NotNil(t, borrowed)
			_, err = borrowed.Stat(".")
			require.Error(t, err, "failed Init must close its pinned root")
			require.NoError(t, other.Cleanup(t.Context()))
			require.NoDirExists(t, other.Root())
		})
	}
}
