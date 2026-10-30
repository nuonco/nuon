package helm

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type replacingChartFS struct {
	afero.Fs
	root, canary, original, phase string
	replaced                      bool
}

func (fs *replacingChartFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	file, err := fs.Fs.OpenFile(name, flag, perm)
	if err != nil || filepath.Base(name) != "Chart.yaml" || flag&os.O_EXCL == 0 {
		return file, err
	}
	return &replacingChartFile{File: file, fs: fs}, nil
}

func (fs *replacingChartFS) replace(name string) error {
	name = filepath.Join(fs.root, name)
	fs.original = name + ".original"
	if err := os.Rename(name, fs.original); err != nil {
		return err
	}
	if err := os.Link(fs.canary, name); err != nil {
		return err
	}
	fs.replaced = true
	return nil
}

type replacingChartFile struct {
	afero.File
	fs *replacingChartFS
}

func (f *replacingChartFile) Fd() uintptr {
	return f.File.(interface{ Fd() uintptr }).Fd()
}

func (f *replacingChartFile) Write(data []byte) (int, error) {
	n, err := f.File.Write(data)
	if err == nil && f.fs.phase == "write" {
		err = f.fs.replace(f.Name())
	}
	return n, err
}

func (f *replacingChartFile) Close() error {
	if err := f.File.Close(); err != nil {
		return err
	}
	if f.fs.phase == "close" && !f.fs.replaced {
		return f.fs.replace(f.Name())
	}
	return nil
}

func TestChartMaterializationTimestampsStayOnOriginalFile(t *testing.T) {
	for _, phase := range []string{"write", "close"} {
		t.Run(phase, func(t *testing.T) {
			parent := t.TempDir()
			root, canary := filepath.Join(parent, "job"), filepath.Join(parent, "canary")
			require.NoError(t, os.Mkdir(root, 0700))
			require.NoError(t, os.WriteFile(canary, []byte("outside canary"), 0640))
			canaryTime, archiveTime := time.Unix(946684800, 0), time.Unix(1700000000, 0)
			require.NoError(t, os.Chtimes(canary, canaryTime, canaryTime))
			pinned, _ := pinnedChartSource(t, root)
			fs := &replacingChartFS{Fs: pinned, root: root, canary: canary, phase: phase}
			data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: acmeChartMetadata})

			dir, err := materializeChartArchive(fs, bytes.NewReader(data))
			require.NoError(t, err)
			require.True(t, fs.replaced, "the real hardlink swap must occur during materialization")
			original, err := os.Stat(fs.original)
			require.NoError(t, err)
			outside, err := os.Stat(canary)
			require.NoError(t, err)
			linked, err := pinned.Stat(filepath.Join(dir, "Chart.yaml"))
			require.NoError(t, err)
			require.True(t, os.SameFile(outside, linked), "replacement must be a hardlink to the outside canary")
			t.Logf("phase=%s archive=%s original=%s outside=%s", phase, archiveTime.UTC(), original.ModTime().UTC(), outside.ModTime().UTC())
			require.True(t, outside.ModTime().Equal(canaryTime), "outside canary timestamp changed")
			require.True(t, original.ModTime().Equal(archiveTime), "archive timestamp must stay on the exclusively created inode")
			contents, err := os.ReadFile(fs.original)
			require.NoError(t, err)
			require.Equal(t, acmeChartMetadata, string(contents))
			contents, err = os.ReadFile(canary)
			require.NoError(t, err)
			require.Equal(t, "outside canary", string(contents))
			require.EqualValues(t, 0640, outside.Mode().Perm())
		})
	}
}

func TestChartMaterializationPreservesMemoryTimestamps(t *testing.T) {
	for name, fs := range map[string]afero.Fs{
		"memory":        afero.NewMemMapFs(),
		"scoped memory": afero.NewBasePathFs(afero.NewMemMapFs(), "/repo"),
	} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, fs.MkdirAll(".", 0700))
			data := chartArchiveFixture(t, chartArchiveEntry{name: "acme/Chart.yaml", data: acmeChartMetadata})
			dir, err := materializeChartArchive(fs, bytes.NewReader(data))
			require.NoError(t, err)
			info, err := fs.Stat(filepath.Join(dir, "Chart.yaml"))
			require.NoError(t, err)
			require.True(t, info.ModTime().Equal(time.Unix(1700000000, 0)), "closing a memory handle must not overwrite the archive timestamp")
			contents, err := afero.ReadFile(fs, filepath.Join(dir, "Chart.yaml"))
			require.NoError(t, err)
			require.Equal(t, acmeChartMetadata, string(contents))
		})
	}
}
