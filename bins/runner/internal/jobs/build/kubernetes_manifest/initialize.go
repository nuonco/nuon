package kubernetes_manifest

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	switch h.state.cfg.SourceType {
	case "inline":
		l.Info("initializing workspace for inline manifest")
		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithWorkspaceID(jobExecution.ID),
		)
		if err != nil {
			l.Error("unable to create workspace for inline manifest", zap.Error(err))
			return err
		}
		h.state.workspace = wkspace

	case "kustomize":
		if h.state.plan.Src == nil || h.state.plan.Src.URL == "" {
			return fmt.Errorf("kustomize source type requires git source")
		}
		l.Info("initializing workspace from git source", zap.String("repo_url", h.state.plan.Src.URL))
		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithNewGitSource(h.state.plan.Src),
			workspace.WithWorkspaceID(jobExecution.ID),
		)
		if err != nil {
			l.Error("unable to create workspace from git source", zap.Error(err))
			return err
		}
		h.state.workspace = wkspace

	default:
		return fmt.Errorf("unsupported source type: %s", h.state.cfg.SourceType)
	}

	if err := h.state.workspace.Init(ctx); err != nil {
		l.Error("unable to initialize workspace", zap.Error(err))
		return fmt.Errorf("unable to initialize workspace: %w", err)
	}

	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch

	return nil
}
