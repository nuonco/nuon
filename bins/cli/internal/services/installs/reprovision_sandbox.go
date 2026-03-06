package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ReprovisionSandbox(ctx context.Context, installID string, asJSON bool) error {
	installID, err := s.selectInstallID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintLn("install id: " + installID)
	}

	workflowID, err := s.api.ReprovisionInstallSandbox(ctx, installID)
	if err != nil {
		return ui.PrintJSONError(err)
	}
	if s.cfg.Debug {
		ui.PrintLn("workflow id: " + workflowID)
	}

	ui.PrintLn("successfully scheduled reprovision of install sandbox")
	if workflowID != "" && s.cfg.Debug {
		ui.PrintLn("workflow id: " + workflowID)
	}

	if workflowID != "" && s.cfg.Preview {
		return s.workflowsTUI(ctx, installID, workflowID)
	}

	return nil
}
