package actions

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateRun(ctx context.Context, installID, actionWorkflowID string, roleName string, asJSON bool) error {
	awc, err := s.api.GetActionWorkflowLatestConfig(ctx, actionWorkflowID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("error getting action workflow config: %w", err))
	}

	req := &models.ServiceCreateInstallActionWorkflowRunRequest{
		ActionWorkflowConfigID: &awc.ID,
		Role:                   roleName,
	}

	if err := s.api.CreateInstallActionWorkflowRun(ctx, installID, req); err != nil {
		return ui.PrintError(fmt.Errorf("error creating action workflow run: %w", err))
	}

	ui.PrintResult(asJSON, fmt.Sprintf("action triggered for action %s", actionWorkflowID), map[string]string{
		"install_id":         installID,
		"action_workflow_id": actionWorkflowID,
		"status":             "triggered",
		"message":            fmt.Sprintf("action triggered for action %s", actionWorkflowID),
	})
	return nil
}
