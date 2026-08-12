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

	if workflowID != "" {
		return s.setCurrentWorkflow(ctx, workflowID, asJSON)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	workflows, _, err := s.listWorkflows(ctx, installID, offset, limit)
	if err != nil {
		return err
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
		return err
	}

	if err := s.setWorkflowID(ctx, selectedWorkflowID); err != nil {
		return err
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
		if asJSON {
			ui.PrintJSON(actionResult{
				InstallID:  installID,
				ID:         selectedWorkflow.ID,
				WorkflowID: selectedWorkflow.ID,
				Status:     "workflow_selected",
			})
		} else {
			s.printWorkflowSetMsg(selectedWorkflow.Name, selectedWorkflow.ID)
		}
	}

	return nil
}

func (s *Service) WorkflowsDeselect(ctx context.Context, asJSON bool) error {
	if err := s.unsetWorkflowID(ctx); err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(actionResult{Status: "workflow_deselected"})
		return nil
	}

	s.printWorkflowUnsetMsg()
	return nil
}

func (s *Service) setCurrentWorkflow(ctx context.Context, workflowID string, asJSON bool) error {
	workflow, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}

	if err := s.setWorkflowID(ctx, workflow.ID); err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(actionResult{
			ID:         workflow.ID,
			WorkflowID: workflow.ID,
			Status:     "workflow_selected",
		})
		return nil
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
	return s.cfg.WriteConfig()
}

func (s *Service) GetWorkflowID() string {
	return s.cfg.GetString("workflow_id")
}

func (s *Service) printWorkflowSetMsg(name, id string) {
	ui.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current workflow is now %s: %s", name, id)))
}

func (s *Service) printWorkflowUnsetMsg() {
	ui.Printf("%s\n", bubbles.InfoStyle.Render("current workflow is now unset"))
}

func (s *Service) printNoWorkflowsMsg() {
	ui.Printf("%s\n", bubbles.BaseStyle.Render("no workflows found for this install"))
}
