package sandbox

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	src := h.state.workspace.Source()

	l.Info("fetching sandbox source files")
	srcFiles, err := h.getSourceFiles(ctx, src.AbsPath())
	if err != nil {
		l.Error("failed to get source files", zap.Error(err))
		h.writeErrorResult(ctx, "fetch files", err)
		return fmt.Errorf("unable to get source files: %w", err)
	}

	l.Info("packing sandbox terraform files into archive")
	if err := h.state.arch.Pack(ctx, l, srcFiles); err != nil {
		l.Error("failed to pack files", zap.Error(err))
		h.writeErrorResult(ctx, "packing files", err)
		return err
	}

	// If no OCI registry destination is configured, the build validates the source
	// and succeeds without pushing an artifact.
	if h.state.regCfg == nil {
		l.Info("no OCI destination configured, skipping push — source validated successfully")
		resultReq := &models.ServiceCreateRunnerJobExecutionResultRequest{
			Success: true,
		}
		if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
			h.errRecorder.Record("write job execution result", err)
		}
		return nil
	}

	l.Info("copying archive to destination", zap.String("dst", h.state.resultTag), zap.Any("cfg", h.state.regCfg))
	res, err := h.ociCopy.CopyFromStore(ctx,
		h.state.arch.Ref(),
		"latest",
		h.state.regCfg,
		h.state.resultTag,
	)
	if err != nil {
		l.Error("failed to copy", zap.Error(err))
		h.writeErrorResult(ctx, "copy image", err)
		return fmt.Errorf("unable to copy image: %w", err)
	}

	l.Info("writing job result")
	resultReq := registry.ToAPIResult(res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
