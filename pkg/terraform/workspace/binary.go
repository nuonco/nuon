package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// LoadBinary installs the binary using the provided binary
func (w *workspace) LoadBinary(ctx context.Context, log hclog.Logger) error {
	if err := w.Binary.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize binary: %w", err)
	}

	execPath, err := w.Binary.Install(ctx, log, w.root)
	if err != nil {
		return fmt.Errorf("unable to get config file from backend: %w", err)
	}
	w.execPath = execPath

	return nil
}
