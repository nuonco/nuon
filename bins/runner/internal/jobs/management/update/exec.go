package update

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/fidiego/systemctl"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/monitor"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("ensuring image config file", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	monitor.EnsureImageConfigFile(ctx, l, h.settings)
	// NOTE(fd): this is run as the root user
	l.Info("restarting", zap.String("systemctlservice.name", monitor.RunnerServiceName), zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	systemctl.Restart(ctx, monitor.RunnerServiceName, systemctl.Options{})

	return nil
}
