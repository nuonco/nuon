package workflow

import (
	"context"
	"os"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithLogger(l),
		workspace.WithGitSource(&plantypes.GitSource{
			URL:  "https://github.com/jonmorehouse/empty",
			Ref:  "main",
			Path: ".",
		}),
		workspace.WithWorkspaceID(jobExecution.ID),
		workspace.WithTmpRoot(h.workspaceRoot(l)),
	)
	if err != nil {
		return err
	}

	h.state.workspace = wkspace
	if err := h.state.workspace.Init(ctx); err != nil {
		return err
	}

	// An image-backed step gets this directory as its bind-mounted working dir
	// and may run as a non-root user, so it has to be writable by any uid. The
	// workspace is created with MkdirAll, whose mode the umask narrows, so the
	// mode is set explicitly here rather than relying on that.
	if h.state.plan != nil && h.state.plan.SourceImage != "" {
		if err := os.Chmod(h.state.workspace.Root(), 0o777); err != nil {
			return errors.Wrap(err, "unable to make action workspace writable by the container")
		}
	}

	return nil
}
