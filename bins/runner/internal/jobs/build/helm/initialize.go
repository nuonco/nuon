package helm

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// create a new workspace here
	wkspace, err := workspace.New(h.v,
		workspace.WithGitSource(h.state.plan.GetWaypointPlan().GetGitSource()),
		workspace.WithWorkspaceID(jobExecution.ID),
	)
	if err != nil {
		return err
	}

	h.state.workspace = wkspace
	if err := h.state.workspace.Init(ctx); err != nil {
		return err
	}

	h.state.arch = ociarchive.New()
	return nil
}
