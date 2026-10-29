package helm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"helm.sh/helm/v4/pkg/chart/loader/archive"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/ignore"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

// Helm's LoadFiles does not apply directory ignore rules, limits, or BOM stripping.
func loadChart(source *sourcefs.FS, chartPath string) (*chartv2.Chart, error) {
	root, _, err := source.Resolve(chartPath)
	if err != nil {
		return nil, err
	}
	info, err := source.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("cannot load irregular chart file %s", chartPath)
		}
		file, err := source.Open(root)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		files, err := chartArchiveFiles(file)
		if err != nil {
			return nil, err
		}
		return loadChartFiles(files)
	}
	if root != "." {
		view, err := source.Sub(root)
		if err != nil {
			return nil, err
		}
		defer view.Close()
		source, root = view, "."
	}

	rules := ignore.Empty()
	ignorePath, _, err := source.ResolveWithin(root, filepath.Join(root, ignore.HelmIgnore))
	if err == nil {
		file, err := source.Open(ignorePath)
		if err != nil {
			return nil, err
		}
		rules, err = ignore.Parse(file)
		file.Close()
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	rules.AddDefaults()

	var files []*archive.BufferedFile
	directories := make(map[string]bool)
	var walk func(string) error
	walk = func(rel string) error {
		name, _, err := source.ResolveWithin(root, filepath.Join(root, rel))
		if err != nil {
			return err
		}
		info, err := source.Lstat(name)
		if err != nil {
			return err
		}
		n := filepath.ToSlash(rel)
		if rel != "." && rules.Ignore(n, info) {
			return nil
		}
		if info.IsDir() {
			if directories[name] {
				return fmt.Errorf("chart directory symlink loop at %s", n)
			}
			directories[name] = true
			defer delete(directories, name)
			entries, err := source.ReadDir(name)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if err := walk(filepath.Join(rel, entry.Name())); err != nil {
					return err
				}
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", n)
		}
		if info.Size() > archive.MaxDecompressedFileSize {
			return fmt.Errorf("chart file %q is larger than the maximum file size %d", info.Name(), archive.MaxDecompressedFileSize)
		}
		file, err := source.Open(name)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(file, archive.MaxDecompressedFileSize+1))
		file.Close()
		if err != nil {
			return fmt.Errorf("error reading %s: %w", n, err)
		}
		if int64(len(data)) > archive.MaxDecompressedFileSize {
			return fmt.Errorf("chart file %q is larger than the maximum file size %d", info.Name(), archive.MaxDecompressedFileSize)
		}
		data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
		files = append(files, &archive.BufferedFile{Name: n, ModTime: info.ModTime(), Data: data})
		return nil
	}
	if err := walk("."); err != nil {
		return nil, err
	}
	return loadChartFiles(files)
}
