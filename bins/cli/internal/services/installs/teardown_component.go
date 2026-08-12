package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
)

func (s *Service) TeardownComponent(ctx context.Context, installID, componentID string, roleName string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	resp, err := s.api.TeardownInstallComponent(ctx, installID, componentID, roleName)
	if err != nil {
		return err
	}

	printActionResult(asJSON, "successfully triggered teardown", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "teardown_triggered",
	})
	return nil
}
