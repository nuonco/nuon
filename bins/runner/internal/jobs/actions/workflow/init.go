package workflow

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithLogger(l),
		workspace.WithGitSource(&plantypes.GitSource{
			URL: "https://github.com/jonmorehouse/empty",
			Ref: "main",
		}),
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
