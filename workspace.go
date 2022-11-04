package terraform

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultWorkspaceFilePerms os.FileMode = 0600
)

type terraformWorkspace interface {
	init(context.Context, *log.Logger) error
	getTfExecPath() string
	getTmpDir() string
	writeFile(string, []byte) error
	cleanup(context.Context) error
}

// workspace represents a local workspace of a terraform module, it's configurations and binaries etc. Inspired by
// terraform cloud workspaces which give you an isolated place to run terraform operations
type workspace struct {
	// ID represents a unique id for the run and is used as part of the tmp dir
	ID     string
	module Module

	// dependencies passed in
	installer     terraformInstaller
	moduleFetcher moduleFetcher

	// internal state
	tmpDir     string
	tfExecPath string
}

func (w *workspace) init(ctx context.Context, l *log.Logger) error {
	tmpDir, err := w.moduleFetcher.createTmpDir(w.ID)
	if err != nil {
		return fmt.Errorf("unable to create tmp dir: %w", err)
	}
	w.tmpDir = tmpDir

	if err = w.installer.initTerraformInstaller(l); err != nil {
		return fmt.Errorf("unable to init terraform installer: %w", err)
	}

	execPath, err := w.installer.installTerraform(ctx)
	if err != nil {
		return fmt.Errorf("unable to install terraform")
	}
	w.tfExecPath = execPath

	if err := w.moduleFetcher.fetchModule(ctx, w.module, w.tmpDir); err != nil {
		return fmt.Errorf("unable to copy module source files: %w", err)
	}

	return nil
}

func (w *workspace) cleanup(ctx context.Context) error {
	if err := w.installer.removeTerraform(ctx); err != nil {
		return fmt.Errorf("unable to remove terraform: %w", err)
	}

	if err := w.moduleFetcher.cleanupTmpDir(w.tmpDir); err != nil {
		return fmt.Errorf("unable to cleanup tmp dir: %w", err)
	}

	return nil
}

func (w *workspace) getTfExecPath() string {
	return w.tfExecPath
}

func (w *workspace) getTmpDir() string {
	return w.tmpDir
}

func (w *workspace) writeFile(filename string, byts []byte) error {
	fp := filepath.Join(w.tmpDir, filename)
	if err := os.WriteFile(fp, byts, defaultWorkspaceFilePerms); err != nil {
		return fmt.Errorf("unable to write workspace file: %w", err)
	}

	return nil
}
