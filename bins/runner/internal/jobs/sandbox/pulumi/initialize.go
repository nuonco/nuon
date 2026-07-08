package pulumi

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if h.state.plan.OCISource != nil {
		l.Info("initializing OCI archive for sandbox provisioning")
		arch := ociarchive.New()
		if err := arch.Initialize(ctx); err != nil {
			return fmt.Errorf("unable to initialize OCI archive: %w", err)
		}

		l.Info("unpacking OCI archive")
		if err := arch.Unpack(ctx, h.state.plan.OCISource.Registry, h.state.plan.OCISource.Tag); err != nil {
			return fmt.Errorf("unable to unpack OCI archive: %w", err)
		}

		h.state.ociArch = arch

		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithWorkspaceID(jobExecution.ID),
			workspace.WithTmpRoot(arch.BasePath()),
		)
		if err != nil {
			return fmt.Errorf("unable to create workspace from OCI source: %w", err)
		}

		h.state.srcWorkspace = wkspace
		if err := h.state.srcWorkspace.Init(ctx); err != nil {
			return fmt.Errorf("unable to init workspace from OCI source: %w", err)
		}
		return nil
	}

	l.Info("initializing source workspace")
	wkspace, err := workspace.New(h.v,
		workspace.WithLogger(l),
		workspace.WithGitSource(h.state.plan.GitSource),
		workspace.WithWorkspaceID(jobExecution.ID),
	)
	if err != nil {
		return err
	}
	h.state.srcWorkspace = wkspace
	if err := h.state.srcWorkspace.Init(ctx); err != nil {
		return err
	}
	return nil
}
