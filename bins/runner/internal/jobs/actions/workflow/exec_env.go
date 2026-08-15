package workflow

import (
	"context"
	"os"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/git"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func (h *handler) createExecEnv(ctx context.Context, l *zap.Logger, src *plantypes.GitSource, cfg *models.AppActionWorkflowStepConfig) error {
	fp := h.outputsFP(cfg)

	// create an empty regular file for outputs, refusing to follow a symlink a
	// prior image-backed action container may have planted at this path. An
	// image-backed step appends to it from inside the container, which may run as
	// a user that can't write a file owned by the runner, so that path needs the
	// wider mode. Host steps keep the tighter one.
	outputsPerm := os.FileMode(0o644)
	if h.state.plan != nil && h.state.plan.SourceImage != "" {
		outputsPerm = 0o666
	}
	if err := safeWriteFile(fp, nil, outputsPerm); err != nil {
		return errors.Wrap(err, "unable to create outputs file")
	}

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
		l.Error("unable to clone repo", zap.Error(err))
		return errors.Wrap(err, "unable to clone repository")
	}

	return nil
}
