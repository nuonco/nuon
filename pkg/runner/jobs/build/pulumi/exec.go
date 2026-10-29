package pulumi

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	source := h.state.source

	l.Info("fetching source files")
	srcFiles, err := ociarchive.SourceFiles(ctx, source, h.state.workspace.Source().AbsPath(), defaultFileType)
	if err != nil {
		l.Error("failed to get source files", zap.Error(err))
		h.writeErrorResult(ctx, "fetch files", err)
		return fmt.Errorf("unable to get source files: %w", err)
	}

	l.Info("packing pulumi files into archive")
	if err := h.state.arch.Pack(ctx, l, source, srcFiles); err != nil {
		l.Error("failed to pack files", zap.Error(err))
		h.writeErrorResult(ctx, "packing files", err)
		return err
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
