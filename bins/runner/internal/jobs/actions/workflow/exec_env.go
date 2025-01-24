package workflow

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
)

func (h *handler) createExecEnv(ctx context.Context, l *zap.Logger, src *git.Source) error {
	if err := git.Clone(ctx, h.state.workspace.Root(), src, l); err != nil {
		return errors.Wrap(err, "unable to clone repository")
	}

	// create file for outputs
	f, err := os.OpenFile(h.outputsFP(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}
	f.Close()

	return nil
}
