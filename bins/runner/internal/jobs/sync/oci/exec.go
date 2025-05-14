package containerimage

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	srcCfg := h.state.plan.Src
	dstCfg := h.state.plan.Dst

	// NOTE(JM): this is ultimately a short cut for now, until we have time to properly handle oci-cleanup jobs.
	//
	// For an OCI Container Image such as a docker-build or oci image this is a noop. For a terraform or helm
	// deploy, this relies on the fact that the previous image oci artifact is still around.
	if job.Operation == models.AppRunnerJobOperationTypeDestroy {
		return nil
	}

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
	h.state.descriptor = res

	resultReq := registry.ToAPIResult(res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
