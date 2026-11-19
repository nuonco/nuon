package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) LoadHooks(ctx context.Context) error {
	if err := w.Hooks.Init(ctx, w.root); err != nil {
		return fmt.Errorf("unable to init hooks: %w", err)
	}

	return nil
}
