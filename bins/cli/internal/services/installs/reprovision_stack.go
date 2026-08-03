package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ReprovisionStack(ctx context.Context, installID string, skipComponents bool, asJSON bool) error {
	installID, err := s.selectInstallID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintLn("install id: " + installID)
	}

	resp, err := s.api.ReprovisionInstallStack(ctx, installID, skipComponents)
	if err != nil {
		return ui.PrintJSONError(err)
	}

	workflowID := workflowIDFromResp(resp)

	if s.cfg.Debug && workflowID != "" {
		ui.PrintLn("workflow id: " + workflowID)
	}

	printActionResult(asJSON, "successfully scheduled reprovision of install stack", actionResult{
		InstallID:  installID,
		WorkflowID: workflowID,
		Status:     "stack_reprovision_scheduled",
	})

	if !asJSON && workflowID != "" && s.cfg.Preview {
		return s.workflowsTUI(ctx, installID, workflowID, false)
	}

	return nil
}
