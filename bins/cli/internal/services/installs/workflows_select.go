package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) WorkflowsSelect(ctx context.Context, installID, workflowID string, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	if workflowID != "" {
		return s.setCurrentWorkflow(ctx, workflowID, asJSON)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	workflows, _, err := s.listWorkflows(ctx, installID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if len(workflows) == 0 {
		s.printNoWorkflowsMsg()
		return nil
	}

	// Convert workflows to selector options
	options := make([]bubbles.WorkflowOption, len(workflows))
	for i, wf := range workflows {
		status := ""
		if wf.Status != nil {
			status = string(wf.Status.Status)
		}
		options[i] = bubbles.WorkflowOption{
			ID:     wf.ID,
			Name:   wf.Name,
			Type:   string(wf.Type),
			Status: status,
		}
	}

	// Show workflow selector
	selectedWorkflowID, err := bubbles.SelectWorkflow(options, s.cfg.Interactive)
	if err != nil {
		return view.Error(err)
	}

	if err := s.setWorkflowID(ctx, selectedWorkflowID); err != nil {
		return view.Error(err)
	}

	// Find selected workflow for display
	var selectedWorkflow *models.AppWorkflow
	for _, wf := range workflows {
		if wf.ID == selectedWorkflowID {
			selectedWorkflow = wf
			break
		}
	}

	if selectedWorkflow != nil {
		s.printWorkflowSetMsg(selectedWorkflow.Name, selectedWorkflow.ID)
	}

	return nil
}

func (s *Service) WorkflowsDeselect(ctx context.Context) error {
	return s.unsetWorkflowID(ctx)
}

func (s *Service) setCurrentWorkflow(ctx context.Context, workflowID string, asJSON bool) error {
	workflow, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	if err := s.setWorkflowID(ctx, workflow.ID); err != nil {
		return err
	}

	s.printWorkflowSetMsg(workflow.Name, workflow.ID)
	return nil
}

func (s *Service) setWorkflowID(ctx context.Context, workflowID string) error {
	s.cfg.Set("workflow_id", workflowID)
	return s.cfg.WriteConfig()
}

func (s *Service) unsetWorkflowID(ctx context.Context) error {
	s.cfg.Set("workflow_id", "")
	fmt.Printf("%s\n", bubbles.InfoStyle.Render("current workflow is now unset"))
	return s.cfg.WriteConfig()
}

func (s *Service) GetWorkflowID() string {
	return s.cfg.GetString("workflow_id")
}

func (s *Service) printWorkflowSetMsg(name, id string) {
	fmt.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current workflow is now %s: %s", name, id)))
}

func (s *Service) printNoWorkflowsMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("no workflows found for this install"))
}
