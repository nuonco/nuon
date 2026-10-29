package ociarchive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

func writeTarLayer(dst io.Writer, source *sourcefs.FS, files []FileRef) error {
	if source == nil {
		return fmt.Errorf("archive source filesystem is required")
	}
	sorted := append([]FileRef(nil), files...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RelPath < sorted[j].RelPath
	})

	gz := gzip.NewWriter(dst)
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
		info, err := source.Lstat(file.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to stat %s: %w", file.AbsPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, _, err := source.Resolve(file.AbsPath)
			if err != nil {
				return fmt.Errorf("unable to resolve symlink target for %s: %w", file.AbsPath, err)
			}
			stat, err := source.Lstat(resolved)
			if err != nil {
				return fmt.Errorf("unable to stat symlink target for %s: %w", file.AbsPath, err)
			}
			if stat.IsDir() {
				link, err := source.Readlink(file.AbsPath)
				if err != nil {
					return fmt.Errorf("unable to read symlink %s: %w", file.AbsPath, err)
				}
				link = filepath.ToSlash(link)
				resolvedLink := path.Join(path.Dir(file.RelPath), link)
				if path.IsAbs(link) || resolvedLink == file.RelPath || validateTarPath(resolvedLink) != nil && resolvedLink != "." {
					return fmt.Errorf("unsafe archive symlink %q -> %q", file.RelPath, link)
				}
				header, err := tar.FileInfoHeader(info, link)
				if err != nil {
					return fmt.Errorf("unable to create symlink header for %s: %w", file.AbsPath, err)
				}
				header.Name = file.RelPath
				header.ModTime = time.Unix(0, 0)
				header.AccessTime = time.Time{}
				header.ChangeTime = time.Time{}
				header.Uid = 0
				header.Gid = 0
				header.Uname = ""
				header.Gname = ""
				header.Format = tar.FormatPAX
				if err := tw.WriteHeader(header); err != nil {
					return fmt.Errorf("unable to write symlink header for %s: %w", file.AbsPath, err)
				}
				continue
			}
			info = stat
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("archive path %q is not a regular file", file.RelPath)
		}

		mode := int64(0644)
		if info.Mode().Perm()&0111 != 0 {
			mode = 0755
		}
		header := &tar.Header{
			Name:     file.RelPath,
			Mode:     mode,
			Size:     info.Size(),
			ModTime:  time.Unix(0, 0),
			Typeflag: tar.TypeReg,
			Format:   tar.FormatPAX,
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %s: %w", file.RelPath, err)
		}

		in, err := source.Open(file.AbsPath)
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
	if name == "" || name == "." || path.IsAbs(name) || path.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") || strings.ContainsAny(name, "\\\x00") || len(name) >= 2 && name[1] == ':' {
		return fmt.Errorf("invalid relative archive path %q", name)
	}
	return nil
}
