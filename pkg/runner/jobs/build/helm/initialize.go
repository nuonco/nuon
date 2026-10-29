package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/rootfs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) (initErr error) {
	defer func() {
		if initErr != nil {
			h.Cleanup(ctx, job, jobExecution)
		}
	}()
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if (h.state.plan.Src == nil || h.state.plan.Src.URL == "") && h.state.cfg.HelmRepoConfig == nil {
		return fmt.Errorf("either source or helm_repo_config must be provided")
	}

	if h.state.plan.Src != nil && h.state.plan.Src.URL != "" {
		l.Info("initializing workspace from git source", zap.String("repo_url", h.state.plan.Src.URL))
		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithFS(h.fs),
			workspace.WithGitSource(h.state.plan.Src),
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
			workspace.WithFS(h.fs),
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
	h.state.source, err = sourcefs.From(h.state.workspace.FS(), h.state.workspace.Root())
	if err != nil {
		return fmt.Errorf("unable to open chart source: %w", err)
	}
	arch := ociarchive.New()
	h.state.arch = arch
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	archiveFS, ok := arch.FS().(*rootfs.FS)
	if !ok {
		return fmt.Errorf("helm archive requires a rooted output filesystem")
	}
	h.state.outputRoot, err = archiveFS.Sub(".")
	if err != nil {
		return fmt.Errorf("unable to open chart output root: %w", err)
	}
	h.state.output, err = sourcefs.From(h.state.outputRoot, h.state.outputRoot.Root())
	if err != nil {
		return fmt.Errorf("unable to open chart output: %w", err)
	}

	return nil
}
