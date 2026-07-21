package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeployComponents(ctx context.Context, installID string, roleName string, planOnly bool, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	resp, err := s.api.DeployInstallComponents(ctx, installID, roleName, planOnly)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	printActionResult(asJSON, "successfully triggered deploy of all install components", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "deploy_triggered",
	})
	return nil
}
