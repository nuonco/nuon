package workflow

import (
	"context"

	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("cleaning up", zap.String("job_type", "actionsworkflow"))
	if h.state.workspace != nil {
		return h.state.workspace.Cleanup(ctx)
	}
	return nil
}
