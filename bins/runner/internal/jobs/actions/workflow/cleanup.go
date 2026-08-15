package workflow

import (
	"context"

	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("cleaning up", zap.String("job_type", "actionsworkflow"))

	// Drops the job's image lease so collection can reclaim the image later. The
	// image itself is left on the host for the next run.
	h.releaseActionImage(jobExecution.ID)

	if h.state.workspace != nil {
		return h.state.workspace.Cleanup(ctx)
	}
	return nil
}
