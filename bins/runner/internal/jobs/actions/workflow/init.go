package workflow

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithLogger(l),
		workspace.WithGitSource(&planv1.GitSource{
			Url: "https://github.com/jonmorehouse/empty",
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
