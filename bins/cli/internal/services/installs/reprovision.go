package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
)

func (s *Service) Reprovision(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	resp, err := s.api.ReprovisionInstall(ctx, installID)
	if err != nil {
		return err
	}

	printActionResult(asJSON, "successfully triggered install reprovision", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "reprovision_triggered",
	})
	return nil
}
