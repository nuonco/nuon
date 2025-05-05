package actions

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *Service) CreateRun(ctx context.Context, installID, actionWorkflowID string, asJSON bool) error {
	awc, err := s.api.GetActionWorkflowLatestConfig(ctx, actionWorkflowID)
	if err != nil {
		ui.PrintError(fmt.Errorf("error getting action workflow config: %w", err))
		return err
	}

	req := &models.ServiceCreateInstallActionWorkflowRunRequest{
		ActionWorkflowConfigID: generics.ToPtr(awc.ID),
	}

	err = s.api.CreateInstallActionWorkflowRun(ctx, installID, req)
	if err != nil {
		ui.PrintError(fmt.Errorf("error creating action workflow run: %w", err))
		return err
	}

	ui.PrintLn(fmt.Sprintf("action triggered for action %s", actionWorkflowID))

	return nil
}
