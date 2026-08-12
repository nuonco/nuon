package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ReprovisionSandbox(ctx context.Context, installID string, skipComponents bool, asJSON bool) error {
	installID, err := s.selectInstallID(ctx, installID)
	if err != nil {
		return err
	}

	if s.cfg.Debug {
		ui.PrintLn("install id: " + installID)
	}

	resp, err := s.api.ReprovisionInstallSandbox(ctx, installID, skipComponents)
	if err != nil {
		return err
	}

	workflowID := ""
	if resp != nil {
		workflowID = resp.WorkflowID
	}

	if s.cfg.Debug && workflowID != "" {
		ui.PrintLn("workflow id: " + workflowID)
	}

	printActionResult(asJSON, "successfully scheduled reprovision of install sandbox", actionResult{
		InstallID:  installID,
		WorkflowID: workflowID,
		Status:     "sandbox_reprovision_scheduled",
	})

	if !asJSON && workflowID != "" && s.cfg.Preview {
		return s.workflowsTUI(ctx, installID, workflowID, false)
	}

	return nil
}
