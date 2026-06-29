package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
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

func (s *Service) TriggerBranchRun(ctx context.Context, appID, branchID string, planOnly, force, asJSON bool) error {
	appID, err := s.resolveAppID(ctx, appID)
	if err != nil {
		return err
	}

	branchID, err = s.selectBranchID(ctx, appID, branchID)
	if err != nil {
		return err
	}

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

	if run.WorkflowID == "" {
		fmt.Printf("Triggered run %s\n", run.ID)
		return nil
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, "", run.WorkflowID, false)
	return nil
}

func (s *Service) ListBranchRuns(ctx context.Context, appID, branchID string, asJSON bool) error {
	appID, err := s.resolveAppID(ctx, appID)
	if err != nil {
		return err
	}

	branchID, err = s.selectBranchID(ctx, appID, branchID)
	if err != nil {
		return err
	}

	workflows, err := s.api.GetAppBranchRuns(ctx, appID, branchID)
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(workflows)
		return nil
	}

	if len(workflows) == 0 {
		fmt.Println("No runs found for this branch.")
		return nil
	}

	options := make([]bubbles.WorkflowOption, len(workflows))
	for i, wf := range workflows {
		status := ""
		if wf.Status != nil {
			status = string(wf.Status.Status)
		}
		options[i] = bubbles.WorkflowOption{
			ID:     wf.ID,
			Name:   wf.CreatedAt,
			Type:   string(wf.Type),
			Status: status,
		}
	}

	selectedID, err := bubbles.SelectWorkflow(options, s.cfg.Interactive)
	if err != nil {
		return err
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, "", selectedID, false)
	return nil
}

func (s *Service) selectBranchID(ctx context.Context, appID, branchID string) (string, error) {
	if branchID != "" {
		return s.resolveAppBranchID(ctx, appID, branchID)
	}

	if !s.cfg.Interactive {
		return "", fmt.Errorf("interactive terminal required for branch selection; use --branch-id flag to specify directly")
	}

	branches, err := s.api.GetAppBranches(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to list app branches: %w", err)
	}
	if len(branches) == 0 {
		return "", fmt.Errorf("no branches found for this app; create one with: nuon apps branches create")
	}

	options := make([]bubbles.BranchOption, len(branches))
	for i, b := range branches {
		options[i] = bubbles.BranchOption{ID: b.ID, Name: b.Name}
	}

	return bubbles.SelectBranch(options, s.cfg.Interactive)
}

func (s *Service) resolveAppID(ctx context.Context, appID string) (string, error) {
	if appID != "" {
		return appID, nil
	}

	configured := s.getAppID()
	if configured != "" {
		return configured, nil
	}

	return "", fmt.Errorf("no app specified; use --app-id or select an app with: nuon config app")
}
