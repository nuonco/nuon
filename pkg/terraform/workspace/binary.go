package workspace

import (
	"context"
	"fmt"
	"os"
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
