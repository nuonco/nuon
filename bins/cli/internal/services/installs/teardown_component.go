package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) TeardownComponent(ctx context.Context, installID, componentID string, roleName string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	resp, err := s.api.TeardownInstallComponent(ctx, installID, componentID, roleName)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	printActionResult(asJSON, "successfully triggered teardown", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "teardown_triggered",
	})
	return nil
}
