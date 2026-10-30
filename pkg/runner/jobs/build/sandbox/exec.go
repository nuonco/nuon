package sandbox

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/rootfs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	terraformbuild "github.com/nuonco/nuon/pkg/runner/jobs/build/terraform"
	"github.com/nuonco/nuon/pkg/runner/jobs/buildinputs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/runner/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	src := h.state.workspace.Source()
	source := h.state.source
	if err := buildinputs.Verify(h.state.plan, source); err != nil {
		return err
	}

	l.Info("fetching sandbox source files")
	srcFiles, err := ociarchive.SourceFiles(ctx, source, src.AbsPath(), defaultFileType)
	if err != nil {
		l.Error("failed to get source files", zap.Error(err))
		h.writeErrorResult(ctx, "fetch files", err)
		return fmt.Errorf("unable to get source files: %w", err)
	}

	if h.state.cfg != nil && h.state.cfg.VendorProviders {
		if _, ok := h.state.workspace.FS().(*rootfs.FS); !ok {
			return fmt.Errorf("terraform provider vendoring requires a disk-backed workspace")
		}
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
		srcFiles, err = ociarchive.SourceFiles(ctx, source, src.AbsPath(), defaultFileType)
		if err != nil {
			l.Error("failed to re-walk source files after provider mirror", zap.Error(err))
			h.writeErrorResult(ctx, "fetch files", err)
			return fmt.Errorf("unable to re-walk source files: %w", err)
		}
	}

	l.Info("packing sandbox terraform files into archive")
	if err := h.state.arch.Pack(ctx, l, source, srcFiles); err != nil {
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
		if err := buildinputs.Report(ctx, h.apiClient, job.ID, jobExecution.ID, resultReq, h.state.plan.SourceIdentity); err != nil {
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
	if err := buildinputs.Report(ctx, h.apiClient, job.ID, jobExecution.ID, resultReq, h.state.plan.SourceIdentity); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
	return nil
}
