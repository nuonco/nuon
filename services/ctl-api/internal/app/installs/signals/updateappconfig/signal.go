package updateappconfig

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appconfigupdated"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "update-app-config"

// branchRunStatusRemovedVersion gates the removal of the branch-run
// config-version status update; in-flight histories scheduled that activity
// and must keep replaying it.
// todo(sk): clean this after terminating old workflows
const branchRunStatusRemovedVersion = "branch-run-version-status-removed-v1"

type Signal struct {
	InstallID      string            `json:"install_id"`
	NewAppConfigID string            `json:"new_app_config_id"`
	DryRun         bool              `json:"dry_run"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	TriggeredBy    string            `json:"triggered_by,omitempty"`

	AppBranchRunID string `json:"app_branch_run_id,omitempty"`
	InstallGroupID string `json:"install_group_id,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	if s.DryRun {
		return signal.SignalLifecycleContext{}
	}

	metadata := map[string]any{}
	if s.Metadata != nil {
		for k, v := range s.Metadata {
			metadata[k] = v
		}
	}

	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "update-app-config",
		Metadata:  metadata,
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.NewAppConfigID == "" {
		return fmt.Errorf("new_app_config_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	diffResult, err := activities.AwaitComputeInstallConfigDiff(ctx, &activities.ComputeInstallConfigDiffInput{
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: s.NewAppConfigID,
	})
	if err != nil {
		return fmt.Errorf("unable to compute config diff: %w", err)
	}

	if s.DryRun {
		return nil
	}

	if err := activities.AwaitUpdateInstallAppConfigID(ctx, &activities.UpdateInstallAppConfigIDInput{
		InstallID:      s.InstallID,
		NewAppConfigID: s.NewAppConfigID,
	}); err != nil {
		return fmt.Errorf("unable to update install app_config_id: %w", err)
	}

	if err := activities.AwaitReconcileInstallComponents(ctx, &activities.ReconcileInstallComponentsInput{
		InstallID: s.InstallID,
	}); err != nil {
		return fmt.Errorf("unable to reconcile install components: %w", err)
	}

	if err := activities.AwaitReconcileInstallActions(ctx, &activities.ReconcileInstallActionsInput{
		InstallID: s.InstallID,
	}); err != nil {
		return fmt.Errorf("unable to reconcile install actions: %w", err)
	}

	if err := activities.AwaitReconcileInstallRunbooks(ctx, &activities.ReconcileInstallRunbooksInput{
		InstallID: s.InstallID,
	}); err != nil {
		return fmt.Errorf("unable to reconcile install runbooks: %w", err)
	}

	metadata := s.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	if s.TriggeredBy != "" {
		metadata["triggered_by"] = s.TriggeredBy
	}

	if s.AppBranchRunID != "" {
		if workflow.GetVersion(ctx, branchRunStatusRemovedVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
			if _, err := activities.AwaitUpdateInstallAppConfigVersionStatus(ctx, &activities.UpdateInstallAppConfigVersionStatusInput{
				AppBranchRunID: s.AppBranchRunID,
				InstallID:      s.InstallID,
				Metadata:       metadata,
			}); err != nil {
				return fmt.Errorf("unable to update install app config version status: %w", err)
			}
		}
	} else {
		if _, err := activities.AwaitCreateInstallAppConfigVersion(ctx, &activities.CreateInstallAppConfigVersionInput{
			InstallID:      s.InstallID,
			OldAppConfigID: install.AppConfigID,
			NewAppConfigID: s.NewAppConfigID,
			Diff:           diffResult.Diff,
			Metadata:       metadata,
		}); err != nil {
			return fmt.Errorf("unable to create install app config version: %w", err)
		}
	}

	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: "install-signals",
		Signal: &appconfigupdated.Signal{
			InstallID:      s.InstallID,
			OldAppConfigID: install.AppConfigID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue appconfig-updated signal: %w", err)
	}

	return nil
}
