package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/browser"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	workflowui "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/workflow"
)

func (s *Service) Workflows(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	workflows, hasMore, err := s.listWorkflows(ctx, installID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflows)
		return nil
	}

	view.RenderPaging(formatWorkflows(workflows), offset, limit, hasMore)
	return nil
}

func formatWorkflows(workflows []*models.AppWorkflow) [][]string {
	data := [][]string{
		{
			"ID",
			"NAME",
			"TYPE",
			"STATUS",
			"STARTED AT",
			"FINISHED AT",
			"UPDATED AT",
		},
	}
	for _, workflow := range workflows {
		startedAt, _ := time.Parse(time.RFC3339Nano, workflow.StartedAt)
		finishedAt, _ := time.Parse(time.RFC3339Nano, workflow.FinishedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, workflow.UpdatedAt)
		status := ""
		if workflow.Status != nil {
			status = string(workflow.Status.Status)
		}

		data = append(data, []string{
			workflow.ID,
			workflow.Name,
			string(workflow.Type),
			status,
			startedAt.Format(time.Stamp),
			finishedAt.Format(time.Stamp),
			updatedAt.Format(time.Stamp),
		})
	}

	return data
}

func (s *Service) listWorkflows(ctx context.Context, appID string, offset, limit int) ([]*models.AppWorkflow, bool, error) {
	workflows, hasMore, err := s.api.GetWorkflows(ctx, appID, &models.GetPaginatedQuery{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return workflows, hasMore, nil
}

// we likely want this to also accept and install id
func (s *Service) WorkflowGet(ctx context.Context, installID, workflowID string) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// only show tui in preview
	if s.cfg.Preview {
		workflowui.App(ctx, s.cfg, s.api, installID, workflowID)
		return nil
	}

	// if not in preview, open workflow in browser
	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		ui.PrintError(err)
	}
	url := fmt.Sprintf("%s/%s/installs/%s/workflows/%s", cfg.DashboardURL, s.cfg.OrgID, installID, workflowID)
	browser.OpenURL(url)
	return nil
}
