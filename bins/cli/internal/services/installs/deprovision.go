package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Deprovision(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	resp, err := s.api.DeprovisionInstall(ctx, installID)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	printActionResult(asJSON, "successfully triggered install deprovision", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "deprovision_triggered",
	})
	return nil
}
