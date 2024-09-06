package docker

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// load access info, workspace source and logger
	src := h.state.workspace.Source()

	// build the image locally, pushing to the local registry
	dockerfile, contextDir, err := h.getBuildContext(
		src,
		h.hcLog,
	)
	if err != nil {
		h.writeErrorResult(ctx, "get build context", err)
		return fmt.Errorf("unable to get build context: %w", err)
	}

	// perform the build
	err = h.buildWithKaniko(ctx, h.hcLog, dockerfile, contextDir, h.state.cfg.BuildArgs)
	if err != nil {
		h.writeErrorResult(ctx, "execute kaniko build", err)
		return fmt.Errorf("unable to execute job: %w", err)
	}

	// copy from the local registry to the destination
	res, err := h.ociCopy.CopyFromLocalRegistry(ctx,
		h.state.resultTag,
		h.state.regCfg,
		h.state.resultTag,
	)
	if err != nil {
		h.writeErrorResult(ctx, "push build", err)
		return fmt.Errorf("unable to copy from runner registry to remote: %w", err)
	}

	// write the api result
	resultReq := registry.ToAPIResult(res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
