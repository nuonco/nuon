package terraform

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("cleaning up terraform workspace", zap.String("path", h.state.tfWorkspace.Root()))
	if err := h.state.tfWorkspace.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup", err)
		l.Info("error cleaning up terraform workspace", zap.Error(err))
	}

	// Workspace cleanup also removes the kyverno-policies directory since
	// policies are now written inside the workspace root.
	l.Info("cleaning up workspace", zap.String("path", h.state.workspace.Root()))
	if err := h.state.workspace.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup", err)
		l.Info("error cleaning up workspace", zap.Error(err))
	}

	h.state = nil
	return nil
}
