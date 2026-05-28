package kubernetes_manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
)

const (
	defaultManifestFilename string = "manifest.yaml"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if h.state.plan.KubernetesManifestDeployPlan.Manifest != "" {
		l.Info("manifest already present in plan, skipping OCI artifact pull")
		return nil
	}

	l.Info("initializing archive...")
	if err := h.state.arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}

	l.Info("unpacking archive...",
		zap.String("repository", h.state.srcCfg.Repository),
		zap.String("tag", h.state.srcTag))
	if err := h.state.arch.Unpack(ctx, h.state.srcCfg, h.state.srcTag); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	manifestPath := filepath.Join(h.state.arch.BasePath(), defaultManifestFilename)

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("invalid archive, manifest.yaml was not found after unpacking: %w", err)
		}
		return fmt.Errorf("error reading manifest after initializing: %w", err)
	}

	manifest := string(manifestBytes)

	// Interpolate {{.nuon.*}} placeholders that survived kustomize unchanged.
	// Inline-manifest deploys never reach this branch (planner pre-renders and
	// the early return above skips the OCI pull), so State is only ever set
	// for the kustomize path. Older planners may not populate it; in that
	// case we just apply the manifest as-is.
	if planState := h.state.plan.KubernetesManifestDeployPlan.State; planState != nil {
		stateMap, err := planState.AsMap()
		if err != nil {
			return fmt.Errorf("unable to flatten install state for kustomize manifest: %w", err)
		}
		rendered, err := render.RenderV2(manifest, stateMap)
		if err != nil {
			return fmt.Errorf("unable to render install state into kustomize manifest: %w", err)
		}
		manifest = rendered
		l.Info("rendered install state into kustomize manifest",
			zap.Int("rendered_size", len(rendered)))
	}

	h.state.plan.KubernetesManifestDeployPlan.Manifest = manifest
	l.Info("manifest loaded from OCI artifact",
		zap.Int("content_size", len(manifest)))

	return nil
}
