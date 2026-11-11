package rootfs_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/rootfs"
)

func open(t *testing.T, path string) *rootfs.FS {
	t.Helper()
	f, err := rootfs.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, f.Close()) })
	return f
}

func TestReadWriteAndLinkMetadata(t *testing.T) {
	f := open(t, t.TempDir())
	require.True(t, filepath.IsAbs(f.Root()))
	require.NoError(t, f.MkdirAll("dir/nested", 0o700))
	require.NoError(t, afero.WriteFile(f, "dir/file", []byte("safe"), 0o600))
	require.NoError(t, f.SymlinkIfPossible("../file", "dir/nested/link"))
	data, err := afero.ReadFile(f, "dir/nested/link")
	require.NoError(t, err)
	require.Equal(t, "safe", string(data))
	info, used, err := f.LstatIfPossible("dir/nested/link")
	require.NoError(t, err)
	require.True(t, used)
	require.NotZero(t, info.Mode()&os.ModeSymlink)
	target, err := f.ReadlinkIfPossible("dir/nested/link")
	require.NoError(t, err)
	require.Equal(t, "../file", target)
	require.NoError(t, afero.WriteFile(f, "dir/nested/link", []byte("updated"), 0o600))
	require.NoError(t, f.Rename("dir/file", "dir/renamed"))
	require.NoError(t, f.Chmod("dir/renamed", 0o700))
	require.NoError(t, f.Chtimes("dir/renamed", time.Now(), time.Now()))
	require.NoError(t, f.RemoveAll("dir/nested"))
	data, err = afero.ReadFile(f, "dir/renamed")
	require.NoError(t, err)
	require.Equal(t, "updated", string(data))
}

func TestRejectTraversalAndEscapingLinks(t *testing.T) {
	parent := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(parent, "job"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(parent, "canary"), []byte("neighbor"), 0o600))
	f := open(t, filepath.Join(parent, "job"))
	for _, name := range []string{"../canary", filepath.Join(parent, "canary"), ""} {
		t.Run(name, func(t *testing.T) {
			_, err := f.Open(name)
			require.Error(t, err)
			require.Error(t, afero.WriteFile(f, name, []byte("bad"), 0o600))
			require.Error(t, f.RemoveAll(name))
			_, err = f.Sub(name)
			require.Error(t, err)
		})
	}
	require.NoError(t, afero.WriteFile(f, "inside", []byte("inside"), 0o600))
	for _, target := range []string{"../canary", filepath.Join(parent, "canary"), filepath.Join(f.Root(), "inside")} {
		require.NoError(t, f.SymlinkIfPossible(target, "link"))
		got, err := f.ReadlinkIfPossible("link")
		require.NoError(t, err)
		require.Equal(t, target, got)
		_, err = afero.ReadFile(f, "link")
		require.Error(t, err)
		require.Error(t, afero.WriteFile(f, "link", []byte("bad"), 0o600))
		require.NoError(t, f.Remove("link"))
	}
	data, err := os.ReadFile(filepath.Join(parent, "canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(data))
}

func TestSubPinsParentHandleAcrossRootReplacement(t *testing.T) {
	parent := t.TempDir()
	job := filepath.Join(parent, "job")
	neighbor := filepath.Join(parent, "neighbor")
	for _, path := range []string{job, neighbor} {
		require.NoError(t, os.MkdirAll(filepath.Join(path, "sub"), 0o700))
	}
	require.NoError(t, os.WriteFile(filepath.Join(neighbor, "sub/canary"), []byte("neighbor"), 0o600))
	f := open(t, job)
	require.NoError(t, afero.WriteFile(f, "sub/canary", []byte("job"), 0o600))
	require.NoError(t, os.Rename(job, job+"-moved"))
	require.NoError(t, os.Symlink(neighbor, job))
	sub, err := f.Sub("sub")
	require.NoError(t, err)
	defer sub.Close()
	require.NoError(t, f.Close())
	data, err := afero.ReadFile(sub, "canary")
	require.NoError(t, err)
	require.Equal(t, "job", string(data))
	require.NoError(t, afero.WriteFile(sub, "canary", []byte("still job"), 0o600))
	data, err = os.ReadFile(filepath.Join(neighbor, "sub/canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(data))
	require.NoError(t, sub.Close())
	_, err = sub.Open("canary")
	require.Error(t, err)
	_, err = f.Sub("sub")
	require.Error(t, err)
}

func TestDirectorySubstitutionCannotRedirectOperations(t *testing.T) {
	parent := t.TempDir()
	job := filepath.Join(parent, "job")
	neighbor := filepath.Join(parent, "neighbor")
	require.NoError(t, os.MkdirAll(filepath.Join(job, "dir"), 0o700))
	require.NoError(t, os.Mkdir(neighbor, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(neighbor, "canary"), []byte("neighbor"), 0o600))
	f := open(t, job)
	require.NoError(t, afero.WriteFile(f, "dir/canary", []byte("job"), 0o600))
	sub, err := f.Sub("dir")
	require.NoError(t, err)
	defer sub.Close()
	require.NoError(t, f.Rename("dir", "saved"))
	require.NoError(t, f.SymlinkIfPossible("../neighbor", "dir"))
	_, err = f.Open("dir/canary")
	require.Error(t, err)
	require.Error(t, afero.WriteFile(f, "dir/canary", []byte("bad"), 0o600))
	require.Error(t, f.Rename("saved/canary", "dir/canary"))
	require.Error(t, f.MkdirAll("dir/new", 0o700))
	require.Error(t, f.RemoveAll("dir/canary"))
	_, err = f.Sub("dir")
	require.Error(t, err)
	require.NoError(t, afero.WriteFile(sub, "canary", []byte("job changed"), 0o600))
	data, err := os.ReadFile(filepath.Join(neighbor, "canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(data))
}

func TestOpenRejectsSymlinkRoot(t *testing.T) {
	link := filepath.Join(t.TempDir(), "link")
	require.NoError(t, os.Symlink(t.TempDir(), link))
	_, err := rootfs.Open(link)
	require.Error(t, err)
}

func TestWorkspacePreparationAndCleanup(t *testing.T) {
	for name, backing := range map[string]afero.Fs{
		"default": nil,
		"os":      afero.NewOsFs(),
		"memory":  afero.NewMemMapFs(),
	} {
		t.Run(name, func(t *testing.T) {
			parentPath := t.TempDir()
			prepare := func(replace bool) afero.Fs {
				t.Helper()
				parent, job, err := rootfs.PrepareFS(backing, parentPath, "job", replace)
				require.NoError(t, err)
				if parent != nil {
					t.Cleanup(func() { require.NoError(t, parent.Close()) })
				}
				if closer, ok := job.(interface{ Close() error }); ok {
					t.Cleanup(func() { require.NoError(t, closer.Close()) })
				}
				return job
			}
			job := prepare(false)
			require.NoError(t, afero.WriteFile(job, "keep", []byte("source"), 0o600))
			data, err := afero.ReadFile(prepare(false), "keep")
			require.NoError(t, err)
			require.Equal(t, "source", string(data))
			job = prepare(true)
			_, err = job.Stat("keep")
			require.True(t, os.IsNotExist(err), "replacement must discard the previous contents")
			info, err := job.Stat(".")
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
			for _, invalid := range []string{".", "..", "../neighbor", "nested/job", filepath.Join(parentPath, "neighbor")} {
				_, _, err := rootfs.PrepareFS(backing, parentPath, invalid, true)
				require.Error(t, err, invalid)
				require.Error(t, rootfs.RemoveFS(backing, parentPath, invalid), invalid)
			}
			require.NoError(t, rootfs.RemoveFS(backing, parentPath, "job"))
			if backing == nil {
				_, err = os.Stat(filepath.Join(parentPath, "job"))
			} else {
				_, err = backing.Stat(filepath.Join(parentPath, "job"))
			}
			require.True(t, os.IsNotExist(err), "cleanup must remove the workspace")
			require.NoError(t, rootfs.RemoveFS(backing, parentPath, "job"), "cleanup must tolerate a missing workspace")
		})
	}
}

func TestWorkspaceSymlinkCannotReplaceOrRemoveNeighbor(t *testing.T) {
	parentPath := t.TempDir()
	neighbor := filepath.Join(parentPath, "neighbor")
	require.NoError(t, os.Mkdir(neighbor, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(neighbor, "canary"), []byte("neighbor"), 0o600))
	jobPath := filepath.Join(parentPath, "job")
	require.NoError(t, os.Symlink(neighbor, jobPath))
	for _, replace := range []bool{false, true} {
		_, _, err := rootfs.PrepareFS(nil, parentPath, "job", replace)
		require.Error(t, err)
	}
	require.NoError(t, rootfs.RemoveFS(nil, parentPath, "job"))
	_, err := os.Lstat(jobPath)
	require.True(t, os.IsNotExist(err))
	data, err := os.ReadFile(filepath.Join(neighbor, "canary"))
	require.NoError(t, err)
	require.Equal(t, "neighbor", string(data))
}

func TestBillyChildLifecycleAndTempNames(t *testing.T) {
	f := open(t, t.TempDir())
	require.NoError(t, f.Mkdir(".git", 0o700))
	b := rootfs.NewBilly(f)
	child, err := b.Chroot(".git")
	require.NoError(t, err)
	file, err := child.TempFile("objects", "pack-")
	require.NoError(t, err)
	require.True(t, filepath.IsLocal(file.Name()))
	require.NoError(t, file.Lock())
	require.NoError(t, file.Unlock())
	_, err = file.Write([]byte("object contents"))
	require.NoError(t, err)
	require.NoError(t, file.Close())
	moved := filepath.Join("objects", "ab", "object")
	require.NoError(t, child.Rename(file.Name(), moved))
	reopened, err := child.Open(moved)
	require.NoError(t, err)
	data, err := io.ReadAll(reopened)
	require.NoError(t, err)
	require.Equal(t, "object contents", string(data))
	require.NoError(t, reopened.Close())
	require.NoError(t, b.Close())
	_, err = child.Open(moved)
	require.Error(t, err)
	_, err = f.Stat(".git")
	require.NoError(t, err, "Billy must not close the borrowed job root")
}
