package installs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type ActionRunOpts struct {
	InstallID string
	ActionID  string
	EnvVars   map[string]string
	Wait      bool
	Timeout   time.Duration
	AsJSON    bool
}

func (s *Service) ActionRun(ctx context.Context, opts ActionRunOpts) error {
	installID, err := lookup.InstallID(ctx, s.api, opts.InstallID)
	if err != nil {
		return ui.PrintError(err)
	}

	// Resolve action name → ID + get latest config.
	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("fetching install: %w", err))
	}

	actions, _, err := s.api.GetActionWorkflows(ctx, install.AppID, nil)
	if err != nil {
		return ui.PrintError(fmt.Errorf("listing actions: %w", err))
	}

	var actionWorkflow *models.AppActionWorkflow
	for _, a := range actions {
		if a.Name == opts.ActionID || a.ID == opts.ActionID {
			actionWorkflow = a
			break
		}
	}
	if actionWorkflow == nil {
		names := make([]string, 0, len(actions))
		for _, a := range actions {
			names = append(names, a.Name)
		}
		return ui.PrintError(fmt.Errorf("action %q not found. available: %s", opts.ActionID, strings.Join(names, ", ")))
	}

	config, err := s.api.GetActionWorkflowLatestConfig(ctx, actionWorkflow.ID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("fetching action config: %w", err))
	}

	// Record the current latest run ID so we can detect when our new run appears.
	// The Create endpoint doesn't return an ID, so we also capture the trigger time
	// and only adopt a run whose CreatedAt is at or after that timestamp.
	recentRuns, _, _ := s.api.GetInstallActionWorkflowRecentRuns(ctx, installID, actionWorkflow.ID, &models.GetPaginatedQuery{Limit: 1})
	var prevRunID string
	if recentRuns != nil && len(recentRuns.Runs) > 0 {
		prevRunID = recentRuns.Runs[0].ID
	}

	// Trigger the run.
	configID := config.ID
	triggerTime := time.Now().Add(-time.Second) // small skew tolerance
	err = s.api.CreateInstallActionWorkflowRun(ctx, installID, &models.ServiceCreateInstallActionWorkflowRunRequest{
		ActionWorkflowConfigID: &configID,
		RunEnvVars:             opts.EnvVars,
	})
	if err != nil {
		return ui.PrintError(fmt.Errorf("triggering action: %w", err))
	}

	if !opts.Wait {
		fmt.Fprintf(os.Stderr, "Triggered %s on %s\n", actionWorkflow.Name, installID)
		return nil
	}

	// Poll until the new run appears and reaches a terminal state.
	fmt.Fprintf(os.Stderr, "Triggered %s, waiting for completion...\n", actionWorkflow.Name)

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		runID            string
		consecutiveErrs  int
		maxConsecutiveErrs = 5
	)
	for {
		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return ui.PrintError(fmt.Errorf("timed out after %s waiting for run to complete", timeout))
			}
			return ui.PrintError(waitCtx.Err())
		case <-time.After(2 * time.Second):
		}

		recentRuns, _, err := s.api.GetInstallActionWorkflowRecentRuns(waitCtx, installID, actionWorkflow.ID, &models.GetPaginatedQuery{Limit: 1})
		if err != nil {
			consecutiveErrs++
			fmt.Fprintf(os.Stderr, "warning: polling failed (%d/%d): %s\n", consecutiveErrs, maxConsecutiveErrs, err)
			if consecutiveErrs >= maxConsecutiveErrs {
				return ui.PrintError(fmt.Errorf("aborting after %d consecutive poll failures: %w", consecutiveErrs, err))
			}
			continue
		}
		consecutiveErrs = 0

		if recentRuns == nil || len(recentRuns.Runs) == 0 {
			continue
		}

		latest := recentRuns.Runs[0]
		// Only adopt a run that is distinct from the pre-trigger latest and
		// whose creation timestamp is at or after our trigger.
		if latest.ID == prevRunID || !runCreatedAfter(latest.CreatedAt, triggerTime) {
			continue
		}

		if runID == "" {
			runID = latest.ID
			fmt.Fprintf(os.Stderr, "Run %s started\n", runID)
		}

		// Action workflow run statuses are defined server-side in
		// app.InstallActionWorkflowRunStatus and are NOT the same vocabulary as
		// models.AppStatus — "finished" is terminal success, not "success".
		switch latest.Status {
		case "finished":
			fmt.Fprintf(os.Stderr, "Run %s completed (%s)\n", runID, latest.Status)
			return s.printActionOutputs(waitCtx, installID, actionWorkflow, opts.AsJSON)
		case "error", "timed-out", "cancelled":
			fmt.Fprintf(os.Stderr, "Run %s failed (%s)\n", runID, latest.StatusDescription)
			return ui.PrintError(fmt.Errorf("action %s: %s", latest.Status, latest.StatusDescription))
		}
	}
}

// runCreatedAfter parses an RFC3339-ish CreatedAt timestamp and reports whether it
// is at or after the trigger time. On parse failure it returns true so we don't get
// stuck waiting on a run we can't timestamp-match.
func runCreatedAfter(createdAt string, trigger time.Time) bool {
	if createdAt == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		t, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return true
		}
	}
	return !t.Before(trigger)
}

func (s *Service) printActionOutputs(ctx context.Context, installID string, action *models.AppActionWorkflow, asJSON bool) error {
	outputs, err := s.api.GetInstallActionWorkflowOutputs(ctx, installID, action.ID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("fetching outputs: %w", err))
	}

	if asJSON {
		ui.PrintJSON(outputs)
		return nil
	}

	if m, ok := outputs.(map[string]any); ok {
		if len(m) == 0 {
			fmt.Fprintln(os.Stderr, "no outputs")
			return nil
		}
		view := ui.NewListView()
		flat := make(map[string]string)
		flattenMap("", m, flat)
		printSection(view, flat)
		return nil
	}

	ui.PrintJSON(outputs)
	return nil
}
