package workflow

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func (h *handler) createExecEnv(ctx context.Context, l *zap.Logger, src *plantypes.GitSource) error {
	// create file for outputs
	f, err := os.OpenFile(h.outputsFP(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}
	f.Close()

	if src == nil || src.URL == "" {
		l.Warn("no connected or public vcs config configured")
		return nil
	}

	dirName := git.Dir(src)
	if h.state.workspace.IsDir(dirName) {
		l.Warn(dirName + " already exists, so removing it")

		if err := h.state.workspace.RmDir(dirName); err != nil {
			return errors.Wrap(err, "unable to cleanup old dir")
		}
	}

	dirPath := h.state.workspace.AbsPath(dirName)
	if err := git.Clone(ctx, dirPath, src, l); err != nil {
		return errors.Wrap(err, "unable to clone repository")
	}

	return nil
}
