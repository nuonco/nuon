package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
	"golang.org/x/sys/unix"
	"helm.sh/helm/v4/pkg/chart/loader/archive"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

func chartComponent(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\:\x00") {
		return fmt.Errorf("invalid chart path component %q", name)
	}
	return nil
}

func chartFilePath(name string) error {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00") || name == "." {
		return fmt.Errorf("invalid chart file path %q", name)
	}
	return nil
}

// Helm's archive loader checks normalized file paths but discards link types and
// skips directory headers. Check the original headers before any publication.
func chartArchiveFiles(in io.ReadSeeker) ([]*archive.BufferedFile, error) {
	z, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	tr := tar.NewReader(z)
	base := ""
	seen := map[string]bool{}
	var total int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch hdr.Typeflag {
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			continue
		}
		name := strings.TrimSuffix(hdr.Name, "/")
		if err := chartFilePath(name); err != nil {
			return nil, err
		}
		parts := strings.SplitN(name, "/", 2)
		if base == "" {
			base = parts[0]
		}
		if parts[0] != base {
			return nil, fmt.Errorf("chart archive contains multiple base directories")
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA && hdr.Typeflag != tar.TypeDir {
			return nil, fmt.Errorf("chart archive contains non-regular entry %q", hdr.Name)
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		if len(parts) != 2 || seen[name] {
			return nil, fmt.Errorf("chart archive contains duplicate or unrooted file %q", name)
		}
		seen[name] = true
		total += hdr.Size
		if hdr.Size < 0 || hdr.Size > archive.MaxDecompressedFileSize || total > archive.MaxDecompressedChartSize {
			return nil, fmt.Errorf("chart archive exceeds Helm decompression limits")
		}
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	files, err := archive.LoadArchiveFiles(in)
	if err != nil {
		return nil, err
	}
	if err := checkChartFiles(files); err != nil {
		return nil, err
	}
	return files, nil
}

func checkChartFiles(files []*archive.BufferedFile) error {
	names := map[string]bool{}
	for _, file := range files {
		if err := chartFilePath(file.Name); err != nil {
			return err
		}
		if names[file.Name] {
			return fmt.Errorf("duplicate chart file %q", file.Name)
		}
		names[file.Name] = true
		// Only charts/<child> at each chart boundary denotes a dependency;
		// files/seed.tgz and other contained assets must remain opaque.
		name := file.Name
		for strings.HasPrefix(name, "charts/") {
			child, rest, nested := strings.Cut(strings.TrimPrefix(name, "charts/"), "/")
			if strings.IndexAny(child, "_.") == 0 {
				break // Helm ignores these subcharts.
			}
			if !nested {
				if strings.HasSuffix(child, ".tgz") {
					if _, err := chartArchiveFiles(bytes.NewReader(file.Data)); err != nil {
						return fmt.Errorf("invalid dependency archive %q: %w", file.Name, err)
					}
				}
				break
			}
			name = rest
		}
	}
	for name := range names {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if names[parent] {
				return fmt.Errorf("chart file %q conflicts with directory %q", name, parent)
			}
		}
	}
	return nil
}

func checkChartPaths(chart *chartv2.Chart) error {
	if err := chartComponent(chart.Name()); err != nil {
		return err
	}
	if err := chartComponent(chart.Metadata.Version); err != nil {
		return err
	}
	for _, dep := range chart.Metadata.Dependencies {
		if err := chartComponent(dep.Name); err != nil {
			return err
		}
	}
	for _, dep := range chart.Dependencies() {
		if err := checkChartPaths(dep); err != nil {
			return err
		}
	}
	return nil
}

func materializeChartArchive(fs afero.Fs, in io.ReadSeeker) (string, error) {
	files, err := chartArchiveFiles(in)
	if err != nil {
		return "", err
	}
	chart, err := loadChartFiles(files)
	if err != nil {
		return "", err
	}
	dir, err := afero.TempDir(fs, ".", ".helm-chart-")
	if err != nil {
		return "", err
	}
	if err := materializeChart(fs, dir, chart); err != nil {
		fs.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// Raw retains Helm's metadata, lockfiles, dependency archives, and timestamps;
// using SaveDir here would reserialize them and omit some of that information.
func materializeChart(fs afero.Fs, dir string, chart *chartv2.Chart) error {
	if err := checkChartPaths(chart); err != nil {
		return err
	}
	for _, file := range chart.Raw {
		if err := chartFilePath(file.Name); err != nil {
			return err
		}
	}
	if err := fs.MkdirAll(dir, 0700); err != nil {
		return err
	}
	for _, file := range chart.Raw {
		name := filepath.Join(dir, filepath.FromSlash(file.Name))
		if err := fs.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return err
		}
		out, err := fs.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, writeErr := out.Write(file.Data)
		handle := out
		for {
			base, ok := handle.(*afero.BasePathFile)
			if !ok {
				break
			}
			handle = base.File
		}
		var timestampErr error
		if writeErr == nil {
			switch handle := handle.(type) {
			case interface{ Fd() uintptr }:
				// A name can be replaced with an outside hardlink even under a pinned root.
				timestamp := unix.NsecToTimeval(file.ModTime.UnixNano())
				timestampErr = unix.Futimes(int(handle.Fd()), []unix.Timeval{timestamp, timestamp})
			case *mem.File:
			default:
				timestampErr = fmt.Errorf("unsupported chart output file %T", handle)
			}
		}
		closeErr := out.Close()
		if err := errors.Join(writeErr, timestampErr, closeErr); err != nil {
			return err
		}
		if memory, ok := handle.(*mem.File); ok {
			// Afero's writable memory handle resets mtime during Close.
			mem.SetModTime(memory.Data(), file.ModTime)
		}
	}
	return nil
}

func loadChartFiles(files []*archive.BufferedFile) (*chartv2.Chart, error) {
	chart, err := loader.LoadFiles(files)
	if err != nil {
		return nil, err
	}
	if err := checkChartPaths(chart); err != nil {
		return nil, err
	}
	return chart, nil
}
