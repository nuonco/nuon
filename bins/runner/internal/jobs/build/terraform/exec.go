package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	src := h.state.workspace.Source()

	l.Info("fetching source files")
	srcFiles, err := h.getSourceFiles(ctx, src.AbsPath())
	if err != nil {
		l.Error("failed to get source files", zap.Error(err))
		h.writeErrorResult(ctx, "fetch files", err)
		return fmt.Errorf("unable to get source files: %w", err)
	}

	if err := h.validateSourceFiles(ctx, srcFiles); err != nil {
		l.Warn("unable to validate terraform build", zap.Error(err))
		// TODO(jm): fail when a validation error happens
	}

	// Optionally vendor providers into the source dir before packing so
	// the resulting OCI artifact carries everything an install runner
	// needs to `terraform init` offline. Gated by a server-side feature
	// flag (`terraform-provider-mirror`) which the planner translates
	// into `cfg.VendorProviders`. The install runner does not look at
	// this flag — it auto-detects the mirror at unpack time.
	if h.state.cfg != nil && h.state.cfg.VendorProviders {
		l.Info("vendoring terraform providers via filesystem mirror")
		opMirrorCtx, endMirror := op.Tool(ctx, "terraform", "provider_mirror")
		err := h.generateProviderMirror(opMirrorCtx, src.AbsPath())
		endMirror(err)
		if err != nil {
			l.Error("failed to generate provider mirror", zap.Error(err))
			h.writeErrorResult(ctx, "vendor providers", err)
			return fmt.Errorf("unable to generate provider mirror: %w", err)
		}

		// Re-walk the source dir so the mirror tree + .terraform.lock.hcl
		// land in the OCI artifact alongside the source files.
		l.Info("re-walking source files after provider mirror")
		srcFiles, err = h.getSourceFiles(ctx, src.AbsPath())
		if err != nil {
			l.Error("failed to re-walk source files after provider mirror", zap.Error(err))
			h.writeErrorResult(ctx, "fetch files", err)
			return fmt.Errorf("unable to re-walk source files: %w", err)
		}
	}

	l.Info("packing terraform files into archive")
	if err := h.state.arch.Pack(ctx, l, srcFiles); err != nil {
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
