package kubernetes_manifest

import (
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
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

	var manifestYAML []byte
	if err := buildinputs.Verify(h.state.plan, h.state.source); err != nil {
		return err
	}

	switch h.state.cfg.SourceType {
	case "inline":
		l.Info("using inline manifest")
		manifestYAML = []byte(h.state.cfg.InlineManifest)

	case "kustomize":
		kustomizePath := filepath.Join(h.state.workspace.Source().AbsPath(), h.state.cfg.KustomizePath)
		l.Info("building kustomize overlay", zap.String("path", kustomizePath))
		_, endK := op.Tool(ctx, "kustomize", "build")
		manifestYAML, err = h.buildKustomization(kustomizePath)
		endK(err)
		if err != nil {
			return fmt.Errorf("kustomize build failed: %w", err)
		}
	}

	l.Info("manifest ready", zap.Int("yaml_size", len(manifestYAML)))

	manifestPath := filepath.Join(h.state.arch.BasePath(), defaultManifestFilename)
	if err := h.writeManifest(manifestYAML); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	srcFiles := []ociarchive.FileRef{
		{
			AbsPath:  manifestPath,
			RelPath:  defaultManifestFilename,
			FileType: defaultFileType,
		},
	}

	l.Info("packing manifest into archive")
	if err := h.state.arch.Pack(ctx, l, h.state.archiveSource, srcFiles); err != nil {
		return fmt.Errorf("unable to pack archive: %w", err)
	}

	l.Info("copying archive to destination")
	res, err := h.ociCopy.CopyFromStore(ctx,
		h.state.arch.Ref(),
		"latest",
		h.state.regCfg,
		h.state.resultTag,
	)
	if err != nil {
		h.errRecorder.Record("copy image", err)
		return err
	}

	l.Info("writing job result")
	resultReq := registry.ToAPIResult(res)
	if err := buildinputs.Report(ctx, h.apiClient, job.ID, jobExecution.ID, resultReq, h.state.plan.SourceIdentity); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}
	return nil
}

func (h *handler) buildKustomization(kustomizePath string) ([]byte, error) {
	opts := krusty.MakeDefaultOptions()
	opts.LoadRestrictions = types.LoadRestrictionsRootOnly
	fs := &kustomizeFS{source: h.state.source}
	root, base, err := fs.CleanedAbs(kustomizePath)
	if err != nil {
		return nil, err
	}
	if base != "" {
		return nil, fmt.Errorf("kustomize path must be a directory: %s", kustomizePath)
	}

	if h.state.cfg.KustomizeConfig != nil {
		if h.state.cfg.KustomizeConfig.EnableHelm {
			return nil, fmt.Errorf("Kustomize enable_helm requires confined subprocess execution and is not supported in rooted builds")
		}
	}

	k := krusty.MakeKustomizer(opts)
	resMap, err := k.Run(fs, root.String())
	if fs.err != nil {
		return nil, fs.err
	}
	if err != nil {
		return nil, err
	}

	return resMap.AsYaml()
}
