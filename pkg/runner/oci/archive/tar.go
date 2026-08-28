package ociarchive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"time"
)

func writeTarLayer(dst string, files []FileRef) (retErr error) {
	sorted := append([]FileRef(nil), files...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RelPath < sorted[j].RelPath
	})

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); retErr == nil {
			retErr = err
		}
	}()

	gz := gzip.NewWriter(out)
	gz.Header.ModTime = time.Unix(0, 0)
	gz.Header.OS = 255
	tw := tar.NewWriter(gz)
	seen := make(map[string]struct{}, len(sorted))
	directories := make(map[string]struct{})
	for _, file := range sorted {
		if err := validateTarPath(file.RelPath); err != nil {
			return err
		}
		if _, ok := seen[file.RelPath]; ok {
			return fmt.Errorf("duplicate archive path %q", file.RelPath)
		}
		seen[file.RelPath] = struct{}{}
		for dir := path.Dir(file.RelPath); dir != "."; dir = path.Dir(dir) {
			directories[dir] = struct{}{}
		}
	}
	directoryNames := make([]string, 0, len(directories))
	for dir := range directories {
		directoryNames = append(directoryNames, dir)
	}
	sort.Strings(directoryNames)
	for _, dir := range directoryNames {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Mode:     0755,
			ModTime:  time.Unix(0, 0),
			Typeflag: tar.TypeDir,
			Format:   tar.FormatPAX,
		}); err != nil {
			return fmt.Errorf("unable to write directory header for %s: %w", dir, err)
		}
	}

	for _, file := range sorted {
		stat, err := os.Stat(file.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to stat %s: %w", file.AbsPath, err)
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("archive path %q is not a regular file", file.RelPath)
		}

		mode := int64(0644)
		if stat.Mode().Perm()&0111 != 0 {
			mode = 0755
		}
		header := &tar.Header{
			Name:     file.RelPath,
			Mode:     mode,
			Size:     stat.Size(),
			ModTime:  time.Unix(0, 0),
			Typeflag: tar.TypeReg,
			Format:   tar.FormatPAX,
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %s: %w", file.RelPath, err)
		}

		in, err := os.Open(file.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to open %s: %w", file.AbsPath, err)
		}
		_, copyErr := io.Copy(tw, in)
		closeErr := in.Close()
		if copyErr != nil {
			return fmt.Errorf("unable to archive %s: %w", file.AbsPath, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("unable to close %s: %w", file.AbsPath, closeErr)
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func validateTarPath(name string) error {
	if name == "" || path.IsAbs(name) || path.Clean(name) != name || name == ".." || len(name) >= 3 && name[:3] == "../" {
		return fmt.Errorf("invalid relative archive path %q", name)
	}
	return nil
}
