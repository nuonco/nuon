package helm

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

	if h.state.plan.Src == nil && h.state.plan.Src.URL != "" && h.state.cfg.HelmRepoConfig == nil {
		return fmt.Errorf("either source or helm_repo_config must be provided")
	}

	if h.state.plan.Src != nil && h.state.plan.Src.URL != "" {
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

	} else {
		l.Info("initializing workspace from helm repo config")
		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithWorkspaceID(jobExecution.ID),
		)
		if err != nil {
			l.Error("unable to create workspace from helm repo config", zap.Error(err))
			return err
		}
		h.state.workspace = wkspace

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
