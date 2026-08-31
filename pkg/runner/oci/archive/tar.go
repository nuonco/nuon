package ociarchive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Unpack keys off this to tell tarball artifacts from pre-tarball ones.
	tarballMediaType string = "application/vnd.nuon.archive.tar+gzip"

	tarballName string = "nuon-archive.tar.gz"
)

func writeTarGz(dst string, files []FileRef) (retErr error) {
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to create tarball: %w", err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)

	for _, f := range files {
		if err := appendTarFile(tw, f); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("unable to close tar writer: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("unable to close gzip writer: %w", err)
	}
	return nil
}

func appendTarFile(tw *tar.Writer, f FileRef) error {
	// Stat, not Lstat: the per-file packer dereferenced symlinks, so we do too.
	stat, err := os.Stat(f.AbsPath)
	if err != nil {
		return fmt.Errorf("unable to stat %s: %w", f.AbsPath, err)
	}
	if stat.IsDir() {
		return nil
	}

	hdr, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return fmt.Errorf("unable to build tar header for %s: %w", f.AbsPath, err)
	}
	hdr.Name = filepath.ToSlash(f.RelPath)

	// Normalized so an unchanged tree hashes to the same layer and the
	// registry skips re-uploading it.
	hdr.ModTime = time.Time{}
	hdr.AccessTime = time.Time{}
	hdr.ChangeTime = time.Time{}
	hdr.Uid, hdr.Gid = 0, 0
	hdr.Uname, hdr.Gname = "", ""

	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("unable to write tar header for %s: %w", f.RelPath, err)
	}

	in, err := os.Open(f.AbsPath)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", f.AbsPath, err)
	}
	defer in.Close()

	if _, err := io.Copy(tw, in); err != nil {
		return fmt.Errorf("unable to write %s into tarball: %w", f.RelPath, err)
	}
	return nil
}

func extractTarGz(src, dstDir string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open tarball: %w", err)
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("unable to read tarball: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("unable to read tar entry: %w", err)
		}

		dst, err := safeJoin(dstDir, hdr.Name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return fmt.Errorf("unable to create %s: %w", hdr.Name, err)
			}
		case tar.TypeReg:
			if err := writeTarEntry(tr, dst, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func writeTarEntry(tr *tar.Reader, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("unable to create dir for %s: %w", dst, err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, tr); err != nil {
		return fmt.Errorf("unable to write %s: %w", dst, err)
	}
	return nil
}

func safeJoin(dir, name string) (string, error) {
	dst := filepath.Join(dir, filepath.FromSlash(name))
	if dst != dir && !strings.HasPrefix(dst, dir+string(os.PathSeparator)) {
		return "", fmt.Errorf("tar entry %q escapes the archive root", name)
	}
	return dst, nil
}
