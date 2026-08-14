package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
)

func (s *Service) RecoverHelmRelease(ctx context.Context, installID, componentID, roleName string, autoApprove, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if !autoApprove {
		confirmed, err := bubbles.ShowConfirmDialog(
			fmt.Sprintf(
				"Recover the Helm release for %s?\n\n"+
					"If an earlier revision rolled out successfully, the release is rolled back to it.\n"+
					"If none ever did, the stuck release is removed so the next deploy can start clean.\n"+
					"Nothing is deployed either way.",
				componentID,
			),
			s.cfg.Interactive,
		)
		if err != nil {
			return ui.PrintError(err)
		}
		if !confirmed {
			return nil
		}
	}

	resp, err := s.api.RecoverInstallComponentHelmRelease(ctx, installID, componentID, roleName)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	printActionResult(asJSON, "successfully triggered helm release recovery", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "helm_release_recovery_triggered",
	})
	return nil
}
