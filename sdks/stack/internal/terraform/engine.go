package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

// installTimeout overrides hc-install's 30s default, which is too short to
// download and unpack the terraform binary (a ~32MB archive that expands to a
// ~110MB binary) on an average connection. When the 30s deadline fires
// mid-download hc-install still tries to chmod the never-written binary,
// surfacing as a confusing "chmod …/terraform: no such file or directory".
const installTimeout = 5 * time.Minute

// resolveTerraform ensures a terraform binary of the requested version is
// available locally and returns its path. An empty version resolves the
// latest stable release. Binaries are cached per version under baseDir so
// multiple installs/versions coexist and re-runs skip the download.
//
// "latest" is cached under its own directory and reused once present — it does
// not auto-upgrade until that cache entry is cleared.
func resolveTerraform(ctx context.Context, ver, baseDir string) (string, error) {
	sub := ver
	if sub == "" {
		sub = "latest"
	}
	dir := filepath.Join(baseDir, sub)
	exe := filepath.Join(dir, binName())
	if fi, err := os.Stat(exe); err == nil && !fi.IsDir() {
		return exe, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create terraform cache dir: %w", err)
	}

	if ver == "" {
		lv := &releases.LatestVersion{
			Product:    product.Terraform,
			InstallDir: dir,
			Timeout:    installTimeout,
		}
		path, err := lv.Install(ctx)
		if err != nil {
			return "", fmt.Errorf("install latest terraform: %w", err)
		}
		return path, nil
	}

	v, err := version.NewVersion(ver)
	if err != nil {
		return "", fmt.Errorf("parse terraform version %q: %w", ver, err)
	}
	ev := &releases.ExactVersion{
		Product:    product.Terraform,
		Version:    v,
		InstallDir: dir,
		Timeout:    installTimeout,
	}
	path, err := ev.Install(ctx)
	if err != nil {
		return "", fmt.Errorf("install terraform %s: %w", ver, err)
	}
	return path, nil
}

func binName() string {
	if runtime.GOOS == "windows" {
		return "terraform.exe"
	}
	return "terraform"
}
