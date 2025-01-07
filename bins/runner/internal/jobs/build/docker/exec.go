package docker

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

	// load access info, workspace source and logger
	src := h.state.workspace.Source()

	// build the image locally, pushing to the local registry
	l.Info("building local context")
	dockerfile, contextDir, err := h.getBuildContext(
		src,
		l,
	)
	if err != nil {
		h.writeErrorResult(ctx, "get build context", err)
		return fmt.Errorf("unable to get build context: %w", err)
	}

	// perform the build
	l.Info("executing build")
	localRef, err := h.buildWithKaniko(ctx, l, dockerfile, contextDir, h.state.cfg.BuildArgs)
	if err != nil {
		h.writeErrorResult(ctx, "execute kaniko build", err)
		return fmt.Errorf("unable to execute job: %w", err)
	}

	l.Info("pushing build to local registry")
	err = h.pushWithKaniko(ctx, l, localRef)
	if err != nil {
		h.writeErrorResult(ctx, "execute kaniko push", err)
		return fmt.Errorf("unable to execute job: %w", err)
	}

	// copy from the local registry to the destination
	l.Info(fmt.Sprintf("copying image from %s to %s", localRef, h.state.resultTag))
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
