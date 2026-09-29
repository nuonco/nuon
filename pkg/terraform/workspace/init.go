package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) InitRoot(ctx context.Context) error {
	if w.root != "" {
		return nil
	}

	if err := w.createRoot(); err != nil {
		return fmt.Errorf("unable to create root: %w", err)
	}

	return nil
}
