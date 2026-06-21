package terraform

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultModuleURL    = "https://github.com/nuonco/install-stacks/archive/refs/heads/main.tar.gz"
	defaultModuleSubdir = "aws"
)

// fetchModule downloads the install-stacks archive over plain HTTPS and
// extracts the requested subdir into destDir. Pure Go (net/http + tar/gzip) so
// it needs no git or other CLI tools — the repo is public, so no auth either.
//
// GitHub archive tarballs nest everything under a single top-level directory
// (e.g. install-stacks-main/); we strip that and the subdir prefix so the
// module's own files (provider.tf, variables.tf, modules/, …) land directly in
// destDir, which terraform then runs as its root configuration.
func fetchModule(ctx context.Context, url, subdir, destDir string) error {
	if url == "" {
		url = defaultModuleURL
	}
	if subdir == "" {
		subdir = defaultModuleSubdir
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build module request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download module %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download module %s: unexpected status %d", url, resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gunzip module archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create module dir: %w", err)
	}

	prefix := subdir + "/"
	extracted := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read module archive: %w", err)
		}

		// Drop the top-level archive directory GitHub adds.
		clean := filepath.ToSlash(filepath.Clean(hdr.Name))
		parts := strings.SplitN(clean, "/", 2)
		if len(parts) < 2 {
			continue
		}
		rel := parts[1]

		// Keep only the requested subdir tree, stripping its prefix.
		if !strings.HasPrefix(rel, prefix) {
			continue
		}
		rel = strings.TrimPrefix(rel, prefix)
		if rel == "" {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(rel))
		// Guard against path traversal in a malicious archive.
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("module archive entry escapes dest dir: %q", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create parent of %s: %w", target, err)
			}
			if err := writeFile(target, tr, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
			extracted++
		}
	}

	if extracted == 0 {
		return fmt.Errorf("module archive %s contained no files under %q", url, subdir)
	}
	return nil
}

func writeFile(path string, r io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
