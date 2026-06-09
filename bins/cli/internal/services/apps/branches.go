package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ListBranches(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	branches, err := s.api.GetAppBranches(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(branches)
		return nil
	}

	data := [][]string{
		{"NAME", "ID", "CREATED"},
	}
	for _, b := range branches {
		data = append(data, []string{
			b.Name,
			b.ID,
			b.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) GetBranch(ctx context.Context, appID, branchID string, asJSON bool) error {
	branch, err := s.api.GetAppBranch(ctx, appID, branchID)
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(branch)
		return nil
	}

	ui.PrintJSON(branch)
	return nil
}

func (s *Service) CreateBranch(ctx context.Context, appID, name string, asJSON bool) error {
	branch, err := s.api.CreateAppBranch(ctx, appID, &models.ServiceCreateAppBranchRequest{
		Name: &name,
	})
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(branch)
		return nil
	}

	fmt.Printf("Created branch %s (%s)\n", branch.Name, branch.ID)
	return nil
}

func (s *Service) TriggerBranchRun(ctx context.Context, appID, branchID string, planOnly, force bool, asJSON bool) error {
	run, err := s.api.TriggerAppBranchRun(ctx, appID, branchID, &models.ServiceTriggerAppBranchRunRequest{
		Force:    force,
		PlanOnly: planOnly,
	})
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(run)
		return nil
	}

	fmt.Printf("Triggered run %s (workflow: %s)\n", run.ID, run.WorkflowID)
	return nil
}

func (s *Service) ListBranchRuns(ctx context.Context, appID, branchID string, asJSON bool) error {
	view := ui.NewListView()

	workflows, err := s.api.GetAppBranchRuns(ctx, appID, branchID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflows)
		return nil
	}

	data := [][]string{
		{"ID", "TYPE", "STATUS", "CREATED"},
	}
	for _, wf := range workflows {
		status := ""
		if wf.Status != nil {
			status = string(wf.Status.Status)
		}
		data = append(data, []string{
			wf.ID,
			string(wf.Type),
			status,
			wf.CreatedAt,
		})
	}
	view.Render(data)
	return nil
}
