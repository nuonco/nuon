package activities

import (
	"context"
	"errors"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/installcomponenthealth"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// installComponentHealthNamespace matches notebookNamespace in
// start_notebook_workflow.go: the installs worker's own namespace, where the
// control-plane-installs queue is served.
const installComponentHealthNamespace = "installs"

type ReconcileUseRunner struct {
	RunnerID string `json:"runner_id" validate:"required"`
	// UseRunner is the use_runner value reported by the install stack outputs.
	// Missing is treated as true by the caller.
	UseRunner bool `json:"use_runner"`
}

// ReconcileUseRunner aligns a runner's status with the use_runner value from
// its install stack outputs. Unlike ReconcileRunnerEnabled, moving to
// not-used is not gated on the runner's current status: a runnerless install
// never goes through the disabled runner_enabled path, so any status except
// disabled can transition straight to not-used. Moving back to a runner only
// applies from not-used, so it can't undo a disabled/re-enabling transition
// handled by ReconcileRunnerEnabled.
//
// Moving to not-used also starts (or confirms running) this install's
// InstallComponentHealthWorkflow, the control-plane stand-in for the Runner's
// own componenthealth.Engine. Moving back to awaiting-heartbeat signals it to
// stop, since a real Runner is about to report the same health itself.
//
// @temporal-gen-v2 activity
// @max-retries 2
// @local
func (a *Activities) ReconcileUseRunner(ctx context.Context, req *ReconcileUseRunner) error {
	if !req.UseRunner {
		if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
			RunnerID:          req.RunnerID,
			Status:            app.RunnerStatusNotUsed,
			StatusDescription: "runner is not used by the install stack",
			Metadata:          map[string]any{app.RunnerOfflineTSMetadataKey: nil},
		}); err != nil {
			return fmt.Errorf("unable to transition runner status: %w", err)
		}

		if err := a.startInstallComponentHealth(ctx, req.RunnerID); err != nil {
			// Health reporting is best-effort: a runnerless install that never
			// gets a health workflow still deploys, it just has no live
			// component health until the next successful reconcile.
			a.l.Warn("unable to start install component health workflow",
				zap.String("runner_id", req.RunnerID), zap.Error(err))
		}
		return nil
	}

	notUsed := app.RunnerStatusNotUsed
	if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          req.RunnerID,
		Status:            app.RunnerStatusAwaitingHeartbeat,
		StatusDescription: "runner was re-enabled by the install stack",
		Metadata:          map[string]any{app.RunnerOfflineTSMetadataKey: nil},
		OnlyIfStatus:      &notUsed,
	}); err != nil {
		return fmt.Errorf("unable to transition runner status: %w", err)
	}

	if err := a.stopInstallComponentHealth(ctx, req.RunnerID); err != nil {
		a.l.Warn("unable to stop install component health workflow",
			zap.String("runner_id", req.RunnerID), zap.Error(err))
	}

	return nil
}

// startInstallComponentHealth starts (or confirms already running)
// InstallComponentHealthWorkflow for runnerID's install. WorkflowID is
// deterministic per install, and WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING
// makes the start idempotent across retries of this activity.
func (a *Activities) startInstallComponentHealth(ctx context.Context, runnerID string) error {
	installID, err := a.installIDForRunner(ctx, runnerID)
	if err != nil {
		return err
	}
	if installID == "" {
		return nil
	}

	opts := tclient.StartWorkflowOptions{
		ID:                       installcomponenthealth.WorkflowID(installID),
		TaskQueue:                installcomponenthealth.TaskQueue,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
	}
	_, err = a.tClient.ExecuteWorkflowInNamespace(ctx, installComponentHealthNamespace, opts,
		"InstallComponentHealthWorkflow", installcomponenthealth.Request{InstallID: installID, RunnerID: runnerID})
	if err != nil {
		return fmt.Errorf("unable to start install component health workflow: %w", err)
	}
	return nil
}

// stopInstallComponentHealth signals InstallComponentHealthWorkflow to stop. A
// workflow that is not running (never started, already stopped, or already
// continued-as-new past its own completion) is not an error here: there is
// nothing left to stop.
func (a *Activities) stopInstallComponentHealth(ctx context.Context, runnerID string) error {
	installID, err := a.installIDForRunner(ctx, runnerID)
	if err != nil {
		return err
	}
	if installID == "" {
		return nil
	}

	err = a.tClient.SignalWorkflowInNamespace(ctx, installComponentHealthNamespace,
		installcomponenthealth.WorkflowID(installID), "", installcomponenthealth.StopSignalName, nil)
	if err != nil && !isWorkflowNotFound(err) {
		return fmt.Errorf("unable to signal install component health workflow to stop: %w", err)
	}
	return nil
}

// installIDForRunner resolves the install owning runnerID, the same lookup
// controlplanejob.injectInstallCredentials makes. A runner belonging to
// anything other than an install's runner group (build/org runners) has no
// install component health workflow.
func (a *Activities) installIDForRunner(ctx context.Context, runnerID string) (string, error) {
	var runner app.Runner
	if err := a.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where(&app.Runner{ID: runnerID}).
		First(&runner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("unable to get runner: %w", err)
	}
	if runner.RunnerGroup.OwnerType != "installs" {
		return "", nil
	}
	return runner.RunnerGroup.OwnerID, nil
}

func isWorkflowNotFound(err error) bool {
	var notFound *serviceerror.NotFound
	return errors.As(err, &notFound)
}
