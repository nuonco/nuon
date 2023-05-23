package workspace

import (
	"context"
	"fmt"
)

// Init: initializes workspace should be called before any other load functions
func (w *workspace) Init(ctx context.Context) error {
	if err := w.createRoot(); err != nil {
		return fmt.Errorf("unable to create root: %w", err)
	}

	return nil
}
