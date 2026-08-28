package terraform

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

		if h.archiveSource != nil {
			store, ref, ok := h.archiveSource.ResolveArchive(h.state.plan.OCISource.Tag)
			if !ok {
				return fmt.Errorf("sandbox source tag %s is not packaged in the bundle; offline runs cannot pull from remote registries", h.state.plan.OCISource.Tag)
			}
			l.Info("unpacking OCI archive from bundle")
			if err := arch.UnpackFromStore(ctx, store, ref); err != nil {
				return fmt.Errorf("unable to unpack OCI archive from bundle: %w", err)
			}
		} else {
			l.Info("unpacking OCI archive")
			if err := arch.Unpack(ctx, h.state.plan.OCISource.Registry, h.state.plan.OCISource.Tag); err != nil {
				return fmt.Errorf("unable to unpack OCI archive: %w", err)
			}
		}

		h.state.ociArch = arch

		// Create a workspace pointed at the unpacked OCI directory (no git clone needed)
		wkspace, err := workspace.New(h.v,
			workspace.WithLogger(l),
			workspace.WithWorkspaceID(jobExecution.ID),
			workspace.WithTmpRoot(arch.BasePath()),
		)
		if err != nil {
			return fmt.Errorf("unable to create workspace from OCI source: %w", err)
		}

		h.state.workspace = wkspace
		if err := h.state.workspace.Init(ctx); err != nil {
			return fmt.Errorf("unable to init workspace from OCI source: %w", err)
		}
		return nil
	}

	l.Info("initializing workspace")
	wkspace, err := workspace.New(h.v,
		workspace.WithLogger(l),
		workspace.WithGitSource(h.state.plan.GitSource),
		workspace.WithWorkspaceID(jobExecution.ID),
	)
	if err != nil {
		return err
	}

	h.state.workspace = wkspace
	if err := h.state.workspace.Init(ctx); err != nil {
		return err
	}
	return nil
}
