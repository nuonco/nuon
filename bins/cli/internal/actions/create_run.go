package actions

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *Service) CreateRun(ctx context.Context, installID, actionWorkflowID string, asJSON bool) error {
	view := ui.NewCreateView("action run", asJSON)
	view.Start()

	awc, err := s.api.GetActionWorkflowLatestConfig(ctx, actionWorkflowID)

	if err != nil {
		return view.Fail(err)
	}

	view.Update("creating action run")
	req := &models.ServiceCreateInstallActionWorkflowRunRequest{
		ActionWorkflowConfigID: generics.ToPtr(awc.ID),
	}

	run, err := s.api.CreateInstallActionWorkflowRun(ctx, installID, req)
	if err != nil {
		return view.Fail(err)
	}

	view.Success(run.ID)

	return nil
}
