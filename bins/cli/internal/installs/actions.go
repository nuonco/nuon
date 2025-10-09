package installs

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	actionui "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/selector"
	// workflowui "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/workflow"
)

func (s *Service) Actions(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	if !s.cfg.Preview {
		return ui.PrintError(errors.New("[NUON_PREVIEW=false] preview is not enabled"))
	}

	// Show workflow selector
	selectedActionWorkflowID, err := selector.ActionSelectorApp(ctx, s.cfg, s.api, s.cfg.AppID, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	actionui.ActionWorkflowApp(ctx, s.cfg, s.api, installID, selectedActionWorkflowID)

	// TODO: execute the action
	// workflowID := ...

	// open the workflow for the action
	// workflowui.WorkflowApp(ctx, s.cfg, s.api, installID, workflowID)
	return nil
}
