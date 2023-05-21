package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) Init(ctx context.Context) error {
	if err := w.createRoot(); err != nil {
		return fmt.Errorf("unable to create root: %w", err)
	}
	return nil
}
