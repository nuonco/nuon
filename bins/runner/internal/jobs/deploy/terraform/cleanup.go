package terraform

import (
	"context"
	"os"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"go.uber.org/zap"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, _ := pkgctx.Logger(ctx)
	l.Info("cleanup: archive")
	if err := h.state.arch.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup archive", err)
	}
	l.Info("cleanup: terraform - removing workspace", zap.String("workspace.root", h.state.tfWorkspace.Root()))
	if err := os.RemoveAll(h.state.tfWorkspace.Root()); err != nil {
		l.Error("cleanup: terraform cleanup failed", zap.String("workspace.root", h.state.tfWorkspace.Root()))
	}
	return nil
}
