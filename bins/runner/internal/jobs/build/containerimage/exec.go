package containerimage

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	srcCfg := h.state.cfg.RepoCfg
	dstCfg := h.state.regCfg

	l.Info(fmt.Sprintf("copying image from %s:%s to %s", h.state.cfg.Image, h.state.cfg.Tag, h.state.plan.DstTag))
	res, err := h.ociCopy.Copy(ctx,
		srcCfg,
		h.state.cfg.Tag,
		dstCfg,
		h.state.resultTag,
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
