package installs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/action/app"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/action/selector"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	// workflowui "github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
)

var errInstallActionsPreviewDisabled = errors.New("[NUON_PREVIEW=false] installs actions is a preview feature, set NUON_PREVIEW=true to enable")

func (s *Service) Actions(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	if !s.cfg.Preview {
		return ui.PrintError(errInstallActionsPreviewDisabled)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// Show workflow selector
	selectedActionWorkflowID, err := selector.App(ctx, s.cfg, s.api, installID, limit, offset)
	if err != nil {
		return ui.PrintError(err)
	}
	app.App(ctx, s.cfg, s.api, installID, selectedActionWorkflowID)

	// TODO: execute the action
	// workflowID := ...

	// open the workflow for the action
	// workflowui.WorkflowApp(ctx, s.cfg, s.api, installID, workflowID)
	return nil
}

func (s *Service) ActionsList(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()
	if !s.cfg.Preview {
		return view.Error(errInstallActionsPreviewDisabled)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return view.Error(err)
	}

	actions, hasMore, err := s.api.GetInstallActionWorkflows(ctx, installID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(actions)
		return nil
	}

	view.RenderPaging(formatInstallActionWorkflows(actions), offset, limit, hasMore)
	return nil
}

func (s *Service) ActionOutputs(ctx context.Context, installID, actionWorkflowID string, asJSON bool) error {
	view := ui.NewGetView()
	if !s.cfg.Preview {
		return view.Error(errInstallActionsPreviewDisabled)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return view.Error(err)
	}

	outputs, err := s.api.GetInstallActionWorkflowOutputs(ctx, installID, actionWorkflowID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(outputs)
		return nil
	}

	b, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return view.Error(err)
	}

	fmt.Println(string(b))
	return nil
}

func formatInstallActionWorkflows(actions []*models.AppInstallActionWorkflow) [][]string {
	data := [][]string{{"NAME", "ACTION ID", "STATUS"}}

	for _, action := range actions {
		name := ""
		actionID := action.ActionWorkflowID
		status := action.Status

		if action.ActionWorkflow != nil {
			name = action.ActionWorkflow.Name
			if actionID == "" {
				actionID = action.ActionWorkflow.ID
			}
			if status == "" {
				status = action.ActionWorkflow.Status
			}
		}

		if name == "" {
			name = actionID
		}
		if actionID == "" {
			actionID = action.ID
		}
		if status == "" {
			status = "-"
		}

		data = append(data, []string{name, actionID, status})
	}

	return data
}
