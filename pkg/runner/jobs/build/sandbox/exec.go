package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	terraformbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/terraform"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/runner/registry"
)

const archivePushAttempts = 3

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

	if h.state.cfg != nil && h.state.cfg.VendorProviders {
		l.Info("vendoring terraform providers via filesystem mirror")
		opVendorCtx, endVendor := op.Tool(ctx, "terraform", "vendor")
		var mirrorPlatforms []string
		if h.cfg != nil {
			mirrorPlatforms = h.cfg.TerraformMirrorPlatforms
		}
		err := terraformbuild.GenerateProviderMirror(opVendorCtx, src.AbsPath(), terraformbuild.MirrorConfig{
			TerraformVersion: h.state.cfg.TerraformVersion,
			MirrorPlatforms:  mirrorPlatforms,
		})
		endVendor(err)
		if err != nil {
			l.Error("failed to generate provider mirror", zap.Error(err))
			h.writeErrorResult(ctx, "vendor providers", err)
			return fmt.Errorf("unable to generate provider mirror: %w", err)
		}

		l.Info("re-walking sandbox source files after provider mirror")
		srcFiles, err = h.getSourceFiles(ctx, src.AbsPath())
		if err != nil {
			l.Error("failed to re-walk source files after provider mirror", zap.Error(err))
			h.writeErrorResult(ctx, "fetch files", err)
			return fmt.Errorf("unable to re-walk source files: %w", err)
		}
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
	var res *ocispec.Descriptor
	for attempt := 1; attempt <= archivePushAttempts; attempt++ {
		res, err = h.ociCopy.CopyFromStore(ctx,
			h.state.arch.Ref(),
			"latest",
			h.state.regCfg,
			h.state.resultTag,
		)
		if err == nil {
			break
		}
		if attempt < archivePushAttempts {
			l.Warn("retrying archive copy after transient registry failure",
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", archivePushAttempts),
				zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * time.Duration(attempt)):
			}
		}
	}
	if err != nil {
		l.Error("failed to copy", zap.Error(err))
		h.writeErrorResult(ctx, "copy image", err)
		return fmt.Errorf("unable to copy image after %d attempts: %w", archivePushAttempts, err)
	}

	l.Info("writing job result")
	resultReq := registry.ToAPIResult(h.state.regCfg.Repository, res)
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
