package terraform

import (
	"context"
	"os"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, _ := pkgctx.Logger(ctx)
	if err := h.state.arch.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup archive", err)
	}

	if err := os.RemoveAll(h.state.tfWorkspace.Root()); err != nil {
		l.Error("cleanup: terraform cleanup failed", zap.String("workspace.root", h.state.tfWorkspace.Root()))
	}
	return nil
}
