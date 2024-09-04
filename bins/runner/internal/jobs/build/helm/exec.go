package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.log.Info("packaging chart")
	packagePath, err := h.packageChart()
	if err != nil {
		return fmt.Errorf("unable to get source files: %w", err)
	}
	h.log.Info("successfully packaged chart", zap.String("path", packagePath))

	h.log.Info("packing chart into archive")
	if err := h.state.arch.Pack(ctx, h.log, []ociarchive.FileRef{
		{
			AbsPath: packagePath,
			RelPath: defaultChartPackageFilename,
		},
	}); err != nil {
		return fmt.Errorf("unable to pack archive with helm archive: %w", err)
	}

	h.log.Info("copying archive to destination")
	res, err := h.ociCopy.CopyFromStore(ctx,
		h.state.arch.Ref(),
		"latest",
		h.state.dstCfg,
		h.state.resultTag,
	)
	if err != nil {
		h.writeErrorResult(ctx, "copy image", err)
		return err
	}

	h.log.Info("writing job result")
	resultReq := registry.ToAPIResult(res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
	return nil
}
