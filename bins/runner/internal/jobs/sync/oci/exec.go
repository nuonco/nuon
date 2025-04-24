package containerimage

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	srcCfg := h.state.plan.Src
	dstCfg := h.state.plan.Dst

	res, err := h.ociCopy.Copy(ctx,
		srcCfg,
		h.state.plan.SrcTag,
		dstCfg,
		h.state.plan.DstTag,
	)
	if err != nil {
		h.writeErrorResult(ctx, "copy image", err)
		return err
	}

	resultReq := registry.ToAPIResult(res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
