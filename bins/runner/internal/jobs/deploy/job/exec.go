package job

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l = l.With(
		zap.String("service.name", "runner.job"),
		zap.String("nuon.tool", "job"),
		zap.String("nuon.deploy.kind", "job"),
	)

	l.Warn("job components are no longer supported, please use an action instead")
	return nil
}
