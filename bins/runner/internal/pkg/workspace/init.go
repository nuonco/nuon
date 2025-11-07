package workspace

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (w *workspace) Init(ctx context.Context) error {
	if err := w.initRootDir(); err != nil {
		w.L.Error("unable to initialize root dir", zap.Error(err))
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
