package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) LoadBinary(ctx context.Context) error {
	if err := w.Binary.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize binary: %w", err)
	}

	// TODO(jm): figure out logging strategy
	execPath, err := w.Binary.Install(ctx, nil, w.root)
	if err != nil {
		return fmt.Errorf("unable to get config file from backend: %w", err)
	}
	w.execPath = execPath

	return nil
}
