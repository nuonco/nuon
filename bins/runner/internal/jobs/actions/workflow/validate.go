package workflow

import (
	"context"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("validating", zap.String("job_type", "actionsworkflow"))
	if err := jobs.Matches(job, h); err != nil {
		return err
	}
	return nil
}
