package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

// LoadBinary installs the binary using the provided binary
func (w *workspace) LoadBinary(ctx context.Context, log hclog.Logger) error {
	if err := w.Binary.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize binary: %w", err)
	}

	installPath := filepath.Join(w.root, "bins")
	if err := os.MkdirAll(installPath, defaultDirPermissions); err != nil {
		return fmt.Errorf("unable to create bins path: %w", err)
	}

	execPath, err := w.Binary.Install(ctx, log, installPath)
	if err != nil {
		return fmt.Errorf("unable to install binary: %w", err)
	}
	w.execPath = execPath

	return nil
}

func (w *workspace) loadLocalBinary(ctx context.Context) {
	// Find the terraform executable path (equivalent to `which terraform`)
	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		panic(err)
	}

	// Ensure the bins directory exists
	err = os.MkdirAll(filepath.Join(w.root, "bins"), 0755)
	if err != nil {
		panic(err)
	}

	// Copy the file
	err = copyFile(terraformPath, filepath.Join(w.root, "/bins/terraform"))
	if err != nil {
		panic(err)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
