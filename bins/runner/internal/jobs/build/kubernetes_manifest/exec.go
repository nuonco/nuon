package kubernetes_manifest

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	var manifestYAML []byte

	switch h.state.cfg.SourceType {
	case "inline":
		l.Info("using inline manifest")
		manifestYAML = []byte(h.state.cfg.InlineManifest)

	case "kustomize":
		kustomizePath := filepath.Join(h.state.workspace.Source().AbsPath(), h.state.cfg.KustomizePath)
		l.Info("building kustomize overlay", zap.String("path", kustomizePath))
		manifestYAML, err = h.buildKustomization(kustomizePath)
		if err != nil {
			return fmt.Errorf("kustomize build failed: %w", err)
		}
	}

	l.Info("manifest ready", zap.Int("yaml_size", len(manifestYAML)))

	manifestPath := filepath.Join(h.state.arch.BasePath(), defaultManifestFilename)
	if err := h.writeManifest(manifestPath, manifestYAML); err != nil {
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
	if err := h.state.arch.Pack(ctx, l, srcFiles); err != nil {
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
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, resultReq); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) buildKustomization(kustomizePath string) ([]byte, error) {
	opts := krusty.MakeDefaultOptions()

	if h.state.cfg.KustomizeConfig != nil {
		switch h.state.cfg.KustomizeConfig.LoadRestrictor {
		case "none":
			opts.LoadRestrictions = types.LoadRestrictionsNone
		default:
			opts.LoadRestrictions = types.LoadRestrictionsRootOnly
		}

		if h.state.cfg.KustomizeConfig.EnableHelm {
			opts.PluginConfig.HelmConfig.Enabled = true
		}
	}

	k := krusty.MakeKustomizer(opts)
	resMap, err := k.Run(filesys.MakeFsOnDisk(), kustomizePath)
	if err != nil {
		return nil, err
	}

	return resMap.AsYaml()
}
