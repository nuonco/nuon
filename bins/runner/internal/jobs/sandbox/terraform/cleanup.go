package terraform

import (
	"context"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
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

	l.Info("cleaning up workspace", zap.String("path", h.state.workspace.Root()))
	if err := h.state.workspace.Cleanup(ctx); err != nil {
		h.errRecorder.Record("unable to cleanup", err)
		l.Info("error cleaning up workspace", zap.Error(err))
	}

	policyDir := filepath.Join("/tmp", h.state.plan.InstallID)
	l.Info("cleaning up policy dir", zap.String("path", policyDir))
	err = os.RemoveAll(policyDir)
	if err != nil {
		h.errRecorder.Record("unable to cleanup policy directory", err)
		l.Info("error cleaning up policy dir", zap.Error(err))
	}

	h.state = nil
	return nil
}
