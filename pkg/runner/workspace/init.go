package workspace

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

func (w *workspace) Init(ctx context.Context) (retErr error) {
	if err := w.validatePaths(); err != nil {
		return err
	}
	if w.fs != nil {
		return fmt.Errorf("workspace already initialized")
	}
	if err := w.initRootDir(); err != nil {
		w.L.Error("unable to initialize root dir", zap.Error(err))
		return fmt.Errorf("unable to initialize root dir: %w", err)
	}
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, w.closeRoot())
		}
	}()

	if !w.isGit() {
		return nil
	}

	if err := w.clone(ctx); err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	return nil
}
