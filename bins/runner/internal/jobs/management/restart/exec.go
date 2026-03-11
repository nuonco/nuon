package restart

import (
	"context"

	"github.com/fidiego/systemctl"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("restarting runner install service",
		zap.String("job_type", "restart"),
		zap.String("service", monitor.RunnerServiceName),
	)
	systemctl.Restart(ctx, monitor.RunnerServiceName, systemctl.Options{})

	return nil
}
