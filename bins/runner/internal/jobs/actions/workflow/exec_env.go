package workflow

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	"go.uber.org/zap"
)

func (h *handler) createExecEnv(ctx context.Context, l *zap.Logger, src *git.Source) error {
	if err := git.Clone(ctx, h.state.workspace.Root(), src, l); err != nil {
		return errors.Wrap(err, "unable to clone repository")
	}

	return nil
}
