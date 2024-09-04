package workspace

import (
	"context"
	"fmt"
)

func (w *workspace) Init(ctx context.Context) error {
	if err := w.initRootDir(); err != nil {
		return fmt.Errorf("unable to initialize root dir: %w", err)
	}

	if !w.isGit() {
		return nil
	}

	if err := w.clone(ctx); err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	return nil
}
