package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

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
