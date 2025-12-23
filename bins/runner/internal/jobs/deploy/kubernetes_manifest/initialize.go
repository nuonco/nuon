package kubernetes_manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
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

	h.state.plan.KubernetesManifestDeployPlan.Manifest = string(manifestBytes)
	l.Info("manifest loaded from OCI artifact",
		zap.Int("content_size", len(manifestBytes)))

	return nil
}
