package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
)

func (s *Service) TeardownComponents(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	resp, err := s.api.TeardownInstallComponents(ctx, installID)
	if err != nil {
		return err
	}

	printActionResult(asJSON, "successfully triggered teardown of all install components", actionResult{
		InstallID:  installID,
		WorkflowID: workflowIDFromResp(resp),
		Status:     "teardown_triggered",
	})
	return nil
}
