package workflow

import (
	"context"

	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/jobs"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
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
