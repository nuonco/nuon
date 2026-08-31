package ociarchive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, rel string, contents string, mode os.FileMode) FileRef {
	t.Helper()

	abs := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0o755))
	require.NoError(t, os.WriteFile(abs, []byte(contents), mode))
	require.NoError(t, os.Chmod(abs, mode))

	return FileRef{AbsPath: abs, RelPath: rel, FileType: "file/terraform"}
}

func TestTarGzRoundTrip(t *testing.T) {
	src := t.TempDir()
	files := []FileRef{
		writeFile(t, src, "main.tf", "resource \"null_resource\" \"a\" {}\n", 0o644),
		writeFile(t, src, ".terraform/modules/eks/main.tf", "module\n", 0o644),
		writeFile(t, src, ".nuon-bin/linux_amd64/terraform", "#!/bin/sh\n", 0o755),
		writeFile(t, src, "empty.tf", "", 0o644),
	}

	tarball := filepath.Join(t.TempDir(), tarballName)
	require.NoError(t, writeTarGz(tarball, files))

	dst := t.TempDir()
	require.NoError(t, extractTarGz(tarball, dst))

	for _, f := range files {
		want, err := os.ReadFile(f.AbsPath)
		require.NoError(t, err)

		got, err := os.ReadFile(filepath.Join(dst, f.RelPath))
		require.NoError(t, err, "%s missing from extracted tarball", f.RelPath)
		require.Equal(t, want, got, f.RelPath)
	}

	stat, err := os.Stat(filepath.Join(dst, ".nuon-bin/linux_amd64/terraform"))
	require.NoError(t, err)
	require.NotZero(t, stat.Mode()&0o111, "executable bit not preserved")
}

func TestSafeJoinRejectsTraversal(t *testing.T) {
	dir := t.TempDir()

	_, err := safeJoin(dir, "../escaped")
	require.Error(t, err)

	_, err = safeJoin(dir, "nested/../ok.tf")
	require.NoError(t, err)
}
