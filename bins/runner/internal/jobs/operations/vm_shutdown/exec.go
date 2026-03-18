package vmshutdown

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	pkgshutdown "github.com/nuonco/nuon/bins/runner/internal/pkg/shutdown"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	shutdownType, ok := job.Metadata["shutdown_type"]
	if ok && shutdownType == "vm" {
		if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
			Status: models.AppRunnerJobStatusFinished,
		}); err != nil {
			return err
		}

		// NOTE(fd): this shuts down so quickly we do lose the tail end of the logs.
		// executes an os shutdown ↴ via dbus w/ a shell fallback w/ a sudo shell fallback
		err = pkgshutdown.Shutdown(ctx, l, h.v)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

// Exec drains all job loops so in-flight jobs (builds, deploys, etc.) can finish
// before the VM is powered off. The actual VM shutdown happens in finishJob(),
// called from Cleanup().
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("draining job loops before VM shutdown",
		zap.Int("loop_count", len(h.jobLoops)),
		zap.Duration("timeout", drainTimeout),
	)
	for _, jl := range h.jobLoops {
		jl.Drain(drainTimeout)
	}

	l.Info("exec", zap.String("job_type", "vm-shutdown"))

	return nil
}
