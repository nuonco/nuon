package installs

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/watch"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
	workflowselector "github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow/selector"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// selectInstallID resolves an install ID by checking the flag, config, or showing
// an interactive selector matching `installs select` behavior.
func (s *Service) selectInstallID(ctx context.Context, installID string) (string, error) {
	if installID == "" {
		installID = s.GetInstallID()
	}
	if installID == "" {
		var (
			installs []*models.AppInstall
			err      error
		)

		if s.cfg.AppID != "" {
			installs, _, err = s.listAppInstalls(ctx, s.cfg.AppID, 0, 50)
		} else {
			installs, _, err = s.listInstalls(ctx, 0, 50)
		}
		if err != nil {
			return "", err
		}
		if len(installs) == 0 {
			return "", fmt.Errorf("no installs found, create one using installs create")
		}

		installOptions := make([]bubbles.InstallOption, len(installs))
		for i, install := range installs {
			installOptions[i] = bubbles.InstallOption{
				ID:   install.ID,
				Name: install.Name,
			}
		}

		selectedID, err := bubbles.SelectInstall(installOptions, s.cfg.Interactive)
		if err != nil {
			return "", err
		}
		installID = selectedID
	}
	return lookup.InstallID(ctx, s.api, installID)
}

func (s *Service) workflowsTUI(ctx context.Context, installID, workflowID string) error {
	installID, err := s.selectInstallID(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// If no workflow ID provided, show selector
	if workflowID == "" {
		selectedID, err := workflowselector.WorkflowSelectorApp(ctx, s.cfg, s.api, installID)
		if err != nil {
			return ui.PrintError(err)
		}
		if selectedID == "" {
			return nil
		}
		workflowID = selectedID
	} else {
		// Validate workflow ID exists
		_, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return ui.PrintError(errors.Wrap(err, "failed to get workflow"))
		}
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, installID, workflowID)
	return nil
}

func (s *Service) workflowsTUILatest(ctx context.Context, installID string) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// Get the latest workflow for this install
	workflows, _, err := s.api.GetWorkflows(ctx, installID, &models.GetPaginatedQuery{Limit: 1, Offset: 0})
	if err != nil {
		return ui.PrintError(errors.Wrap(err, "failed to get workflows"))
	}

	if len(workflows) == 0 {
		return ui.PrintError(errors.New("no workflows found for this install"))
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, installID, workflows[0].ID)
	return nil
}

// WorkflowsWatchTUI launches the full-screen TUI for watching all workflows for an install.
// It accepts either an installID or workflowID. If workflowID is provided, it resolves
// the install ID from the workflow's OwnerID field.
// Returns an exit code and error for proper CLI exit handling.
func (s *Service) WorkflowsWatchTUI(ctx context.Context, installID, workflowID string) (int, error) {
	var resolvedInstallID string

	// If workflow ID provided, resolve install ID from workflow
	if workflowID != "" {
		workflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return ExitCodeFailed, ui.PrintError(errors.Wrap(err, "failed to get workflow"))
		}
		if workflow.OwnerType != "installs" {
			return ExitCodeFailed, ui.PrintError(errors.Newf("workflow %s is not owned by an install (owner_type: %s)", workflowID, workflow.OwnerType))
		}
		resolvedInstallID = workflow.OwnerID
	} else if installID != "" {
		var err error
		resolvedInstallID, err = lookup.InstallID(ctx, s.api, installID)
		if err != nil {
			return ExitCodeFailed, ui.PrintError(err)
		}
	} else {
		var err error
		resolvedInstallID, err = s.selectInstallID(ctx, installID)
		if err != nil {
			return ExitCodeFailed, ui.PrintError(err)
		}
	}

	exitCode := watch.WatchApp(ctx, s.cfg, s.api, resolvedInstallID)
	return exitCode, nil
}
