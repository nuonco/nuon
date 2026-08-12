package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
)

func (s *Service) DeprovisionSandbox(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	resp, err := s.api.DeprovisionInstallSandbox(ctx, installID)
	if err != nil {
		return err
	}

	printActionResult(asJSON, "successfully scheduled deprovision of install sandbox", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "sandbox_deprovision_scheduled",
	})
	return nil
}
