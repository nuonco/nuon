package installconfigsync

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "install-config-sync"

type Signal struct {
	InstallID   string `json:"install_id" validate:"required"`
	InstallName string `json:"install_name,omitempty"`

	AppInstallConfigSyncID string `json:"app_install_config_sync_id,omitempty"`

	AppBranchID       string `json:"app_branch_id,omitempty"`
	AppBranchConfigID string `json:"app_branch_config_id,omitempty"`
	AppBranchRunID    string `json:"app_branch_run_id,omitempty"`

	CommitSHA   string `json:"commit_sha,omitempty"`
	TriggeredBy string `json:"triggered_by"`
	SourceDir   string `json:"source_dir,omitempty"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "install-config-sync",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		OwnerName: s.InstallName,
		Metadata: map[string]any{
			"install_name": s.InstallName,
			"triggered_by": s.TriggeredBy,
			"commit_sha":   s.CommitSHA,
		},
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("install not found: %w", err)
	}

	if s.InstallName == "" {
		s.InstallName = install.Name
	}

	return nil
}
