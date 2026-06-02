package pulumi

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
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
